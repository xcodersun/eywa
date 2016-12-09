package models

import (
	"errors"
	"fmt"
	"gopkg.in/olivere/elastic.v3"
	. "github.com/eywa/configs"
	. "github.com/eywa/utils"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"time"
)

var ScrollSize = 2000
var KeepAlive = "5m"

type RawQuery struct {
	Channel   *Channel
	Tags      map[string]string
	TimeStart time.Time
	TimeEnd   time.Time
	Nop       bool
}

func (q *RawQuery) Parse(params map[string]string) error {
	if timeRange, found := params["time_range"]; found {
		ranges := strings.Split(timeRange, ":")
		if len(ranges) != 2 || len(ranges[0]) == 0 {
			return errors.New("invalid time_range format: " + timeRange)
		}
		start, err := strconv.ParseInt(ranges[0], 10, 64)
		if err != nil {
			return errors.New("error parsing time_range: " + timeRange)
		}
		q.TimeStart = time.Unix(MilliSecToSec(start), MilliSecToNano(start)).UTC()

		if len(ranges[1]) > 0 {
			end, err := strconv.ParseInt(ranges[1], 10, 64)
			if err != nil {
				return errors.New("error parsing time_range: " + timeRange)
			}
			q.TimeEnd = time.Unix(MilliSecToSec(end), MilliSecToNano(end)).UTC()
		} else {
			q.TimeEnd = time.Now().UTC()
		}
	} else {
		return errors.New("missing time_range")
	}

	if !q.TimeStart.IsZero() && q.TimeStart.After(q.TimeEnd) {
		return errors.New("invalid time_range, start_time is later than end_time")
	}

	if nop, found := params["nop"]; found {
		if nop == "false" {
			q.Nop = false
		} else {
			q.Nop = true
		}
	} else {
		q.Nop = true
	}

	q.Tags = make(map[string]string)
	if tagStr, found := params["tags"]; found {
		tags := strings.Split(tagStr, ",")
		for _, tag := range tags {
			t := strings.Split(tag, ":")
			if len(t) != 3 {
				return errors.New("error parsing tagging: " + tag)
			} else if t[1] != "eq" {
				return errors.New("unsupported operator for tagging: " + t[1])
			} else if len(t[2]) == 0 {
				return errors.New("empty tagging value: " + tag)
			} else if !StringSliceContains(q.Channel.Tags, t[0]) && !StringSliceContains(InternalTags, t[0]) {
				return errors.New("undefined tag: " + t[0] + " on channel: " + q.Channel.Name)
			} else {
				q.Tags[t[0]] = t[2]
			}
		}
	}

	return nil
}

func (q *RawQuery) QueryES() (interface{}, error) {
	res := map[string]interface{}{}

	indexName := TimedIndices(q.Channel, q.TimeStart, q.TimeEnd)
	if len(indexName) == 0 {
		res["size"] = "0b"
		return res, nil
	}

	boolQ := elastic.NewBoolQuery()

	termQs := make([]elastic.Query, 0)
	for tagN, tagV := range q.Tags {
		termQs = append(termQs, elastic.NewTermQuery(tagN, tagV))
	}
	boolQ.Must(termQs...)

	rangeQ := elastic.NewRangeQuery("timestamp").
		From(NanoToMilli(q.TimeStart.UnixNano())).
		To(NanoToMilli(q.TimeEnd.UnixNano()))
	boolQ.Must(rangeQ)

	if q.Nop {
		filterAgg := elastic.NewFilterAggregation()
		filterAgg.Filter(boolQ).SubAggregation("bytes", elastic.NewSumAggregation().Field("_size"))
		resp, err := IndexClient.Search().
			SearchType("count").
			Index(indexName).
			Type(IndexTypeMessages).
			Aggregation("name", filterAgg).
			Do()

		if err != nil {
			return nil, err
		}

		aggs, success := resp.Aggregations.Filter("name")
		if !success {
			return nil, errors.New("error query raw data")
		}
		sum, success := aggs.Sum("bytes")
		if !success {
			return nil, errors.New("error query raw data")
		}

		bytes := int64(*sum.Value)
		if bytes < 1024 {
			res["size"] = fmt.Sprintf("%db", bytes)
		} else if bytes < 1024*1024 {
			res["size"] = fmt.Sprintf("%dkb", bytes/1024)
		} else if bytes < 1024*1024*1024 {
			res["size"] = fmt.Sprintf("%dmb", bytes/1024/1024)
		} else {
			res["size"] = fmt.Sprintf("%dgb", bytes/1024/1024/1024)
		}

		return res, nil

	} else {
		tmpFile, err := ioutil.TempFile(os.TempDir(), fmt.Sprintf("%s.raw", q.Channel.Name))
		if err != nil {
			return nil, err
		}

		size := ScrollSize / Config().Indices.NumberOfShards
		resp, err := IndexClient.Scroll().KeepAlive(KeepAlive).Index(indexName).Type(IndexTypeMessages).Query(boolQ).Size(size).Do()
		for err == nil {
			resp, err = IndexClient.Scroll().KeepAlive(KeepAlive).ScrollId(resp.ScrollId).GetNextPage()
			if err == nil {
				for _, hit := range resp.Hits.Hits {
					tmpFile.Write([]byte(*hit.Source))
					tmpFile.WriteString("\n")
				}
			}
		}

		if err != nil && err != elastic.EOS {
			return nil, err
		}

		res["file"] = tmpFile.Name()
		return res, nil
	}
}
