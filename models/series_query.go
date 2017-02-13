package models

import (
	"errors"
	"gopkg.in/olivere/elastic.v3"
	. "github.com/eywa/utils"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var SeriesAggName = "series_agg"

type SeriesQuery struct {
	Channel      *Channel
	Device       string
	Field        string
	Tags         map[string]string
	SummaryType  string
	TimeStart    time.Time
	TimeEnd      time.Time
	TimeInterval string
}

func (q *SeriesQuery) Parse(params map[string]string) error {
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
		if q.SummaryType == "last" || !StringSliceContains(SupportedSummaryTypes, q.SummaryType) {
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
	} else {
		return errors.New("missing time_range")
	}

	if q.TimeStart.After(q.TimeEnd) {
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

	if itv, found := params["time_interval"]; found {
		if matched, _ := regexp.MatchString(`\d+[yMwdhms]`, itv); !matched {
			return errors.New("invalid time_interval format: " + itv)
		} else {
			q.TimeInterval = itv
		}
	} else {
		return errors.New("missing time_interval")
	}

	return nil
}

func (q *SeriesQuery) QueryES() (interface{}, error) {
	series := make([]map[string]interface{}, 0)
	indexName := TimedIndices(q.Channel, q.TimeStart, q.TimeEnd)
	if len(indexName) == 0 {
		return series, nil
	}

	filterAgg := elastic.NewFilterAggregation()

	boolQ := elastic.NewBoolQuery()

	if q.Device != "" {
		termQ := elastic.NewTermQuery("device_id", q.Device)
		boolQ.Must(termQ)
	}

	termQs := make([]elastic.Query, 0)
	for tagN, tagV := range q.Tags {
		termQs = append(termQs, elastic.NewTermQuery(tagN, tagV))
	}
	boolQ.Must(termQs...)

	rangeQ := elastic.NewRangeQuery("timestamp").
		From(NanoToMilli(q.TimeStart.UnixNano())).
		To(NanoToMilli(q.TimeEnd.UnixNano()))
	boolQ.Must(rangeQ)

	filterAgg.Filter(boolQ)

	var agg elastic.Aggregation
	switch q.SummaryType {
	case "sum":
		agg = elastic.NewDateHistogramAggregation().
			Field("timestamp").
			Interval(q.TimeInterval).
			SubAggregation(q.SummaryType, elastic.NewSumAggregation().Field(q.Field))
	case "avg":
		agg = elastic.NewDateHistogramAggregation().
			Field("timestamp").
			Interval(q.TimeInterval).
			SubAggregation(q.SummaryType, elastic.NewAvgAggregation().Field(q.Field))
	case "min":
		agg = elastic.NewDateHistogramAggregation().
			Field("timestamp").
			Interval(q.TimeInterval).
			SubAggregation(q.SummaryType, elastic.NewMinAggregation().Field(q.Field))
	case "max":
		agg = elastic.NewDateHistogramAggregation().
			Field("timestamp").
			Interval(q.TimeInterval).
			SubAggregation(q.SummaryType, elastic.NewMaxAggregation().Field(q.Field))
	}

	filterAgg.SubAggregation(SeriesAggName, agg)

	resp, err := IndexClient.Search().
		SearchType("count").
		Index(indexName).
		Type(IndexTypeMessages).
		Aggregation("name", filterAgg).
		Do()
	if err != nil {
		return nil, err
	}
	filteredResp, success := resp.Aggregations.Filter("name")
	if !success {
		return nil, errors.New("error querying indices")
	}
	SeriesResp, success := filteredResp.Aggregations.DateHistogram(SeriesAggName)
	if !success {
		return nil, errors.New("error querying indices")
	}

	for _, bkt := range SeriesResp.Buckets {
		switch q.SummaryType {
		case "sum":
			if sum, found := bkt.Aggregations.Sum(q.SummaryType); found {
				series = append(series, map[string]interface{}{"timestamp": bkt.Key, "value": sum.Value})
			}
		case "avg":
			if avg, found := bkt.Aggregations.Sum(q.SummaryType); found {
				series = append(series, map[string]interface{}{"timestamp": bkt.Key, "value": avg.Value})
			}
		case "min":
			if min, found := bkt.Aggregations.Sum(q.SummaryType); found {
				series = append(series, map[string]interface{}{"timestamp": bkt.Key, "value": min.Value})
			}
		case "max":
			if max, found := bkt.Aggregations.Sum(q.SummaryType); found {
				series = append(series, map[string]interface{}{"timestamp": bkt.Key, "value": max.Value})
			}
		}
	}

	return series, nil
}
