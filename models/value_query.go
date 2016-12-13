package models

import (
	"errors"
	"fmt"
	"gopkg.in/olivere/elastic.v3"
	. "github.com/eywa/utils"
	"strconv"
	"strings"
	"time"
)

var SupportedSummaryTypes = []string{"avg", "min", "max", "sum", "last"}
var SupportedOperators = []string{"eq", "ne", "lt", "gt", "le", "ge"}
var ValueAggName = "value_agg"

type ValueQuery struct {
	Channel     *Channel
	Field       string
	Tags        map[string]string
	SummaryType string
	TimeStart   time.Time
	TimeEnd     time.Time
}

func (q *ValueQuery) Parse(params map[string]string) error {
	if field, found := params["field"]; !found {
		return errors.New("missing field")
	} else {
		q.Field = field
		if _, found := q.Channel.Fields[q.Field]; !found {
			return errors.New("undefined field: " + q.Field + " on channel: " + q.Channel.Name)
		}
	}

	if sum, found := params["summary_type"]; !found {
		return errors.New("missing summary_type")
	} else {
		q.SummaryType = sum
		if !StringSliceContains(SupportedSummaryTypes, q.SummaryType) {
			return errors.New("unsupported summary_type: " + q.SummaryType)
		}
	}

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
	}

	if q.SummaryType != "last" && q.TimeStart.IsZero() {
		return errors.New("missing time_range for summary_type: " + q.SummaryType)
	}

	if !q.TimeStart.IsZero() && q.TimeStart.After(q.TimeEnd) {
		return errors.New("invalid time_range, start_time is later than end_time")
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

func TimedIndices(ch *Channel, tStart, tEnd time.Time) string {
	allIndices := ch.Indices()

	indices := []string{}
	oneWeek := 7 * 24 * time.Hour
	t := tStart
	for {
		index := TimedIndexName(ch, t)
		if StringSliceContains(allIndices, index) {
			indices = append(indices, index)
		}

		y, w := t.ISOWeek()
		ey, ew := tEnd.ISOWeek()
		if y > ey || (y == ey && w >= ew) {
			break
		} else {
			t = t.Add(oneWeek)
		}
	}
	return strings.Join(indices, ",")
}

func GlobalIndexName(ch *Channel) string {
	return fmt.Sprintf("channels.%d.*", ch.Id)
}

func (q *ValueQuery) QueryES() (interface{}, error) {
	filterAgg := elastic.NewFilterAggregation()

	boolQ := elastic.NewBoolQuery()

	termQs := make([]elastic.Query, 0)
	for tagN, tagV := range q.Tags {
		termQs = append(termQs, elastic.NewTermQuery(tagN, tagV))
	}
	boolQ.Must(termQs...)

	existsQs := elastic.NewExistsQuery(q.Field)
	boolQ.Must(existsQs)

	if !q.TimeStart.IsZero() {
		rangeQ := elastic.NewRangeQuery("timestamp").
			From(NanoToMilli(q.TimeStart.UnixNano())).
			To(NanoToMilli(q.TimeEnd.UnixNano()))
		boolQ.Must(rangeQ)
	}

	filterAgg.Filter(boolQ)

	var agg elastic.Aggregation
	switch q.SummaryType {
	case "sum":
		agg = elastic.NewSumAggregation().Field(q.Field)
	case "avg":
		agg = elastic.NewAvgAggregation().Field(q.Field)
	case "min":
		agg = elastic.NewMinAggregation().Field(q.Field)
	case "max":
		agg = elastic.NewMaxAggregation().Field(q.Field)
	}

	filterAgg.SubAggregation(ValueAggName, agg)

	if q.SummaryType != "last" {
		indexName := TimedIndices(q.Channel, q.TimeStart, q.TimeEnd)
		if len(indexName) == 0 {
			return nil, nil
		}

		resp, err := IndexClient.Search().
			SearchType("count").
			Index(indexName).
			Type(IndexTypeMessages).
			Aggregation("name", filterAgg).
			Do()
		if err != nil {
			return nil, err
		}
		filterAggResp, success := resp.Aggregations.Filter("name")
		if !success {
			return nil, errors.New("error querying indices")
		}

		var statsResp *elastic.AggregationValueMetric
		switch q.SummaryType {
		case "sum":
			statsResp, success = filterAggResp.Aggregations.Sum(ValueAggName)
		case "avg":
			statsResp, success = filterAggResp.Aggregations.Avg(ValueAggName)
		case "min":
			statsResp, success = filterAggResp.Aggregations.Min(ValueAggName)
		case "max":
			statsResp, success = filterAggResp.Aggregations.Max(ValueAggName)
		}

		if !success {
			return nil, errors.New("error querying indices")
		}

		return map[string]interface{}{"value": statsResp.Value}, nil
	} else {
		resp, err := IndexClient.Search().
			Index(GlobalIndexName(q.Channel)).
			Type(IndexTypeMessages).
			FetchSource(false).
			Field(q.Field).
			Query(boolQ).
			Sort("timestamp", false).
			From(0).Size(1).
			Do()

		if err != nil {
			return nil, err
		}
		if resp.TotalHits() == 0 || resp.Hits == nil ||
			len(resp.Hits.Hits) == 0 {
			return nil, nil
		} else if _, found := resp.Hits.Hits[0].Fields[q.Field]; !found {
			return nil, nil
		} else {
			values, ok := resp.Hits.Hits[0].Fields[q.Field].([]interface{})
			if !ok || len(values) == 0 {
				return nil, nil
			} else {
				return map[string]interface{}{"value": values[0]}, nil
			}
		}
	}
}
