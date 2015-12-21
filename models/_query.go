package models

import (
	"errors"
	"gopkg.in/olivere/elastic.v3"
	"strconv"
	"strings"
	"time"
)

var SupportedSummaryTypes = []string{"avg", "min", "max", "last"}
var SupportedOperators = []string{"eq", "ne", "lt", "gt", "le", "ge"}

type Tag struct {
	Name     string
	Operator string
	Value    string
}

func (t *Tag) ParseTag(tagStr string) error {
	ct := strings.Split(tagStr, ":")
	if len(cr) != 3 || len(cr[0]) == 0 || len(cr[1]) == 0 || len(cr[2]) == 0 {
		return errors.New("error parsing tags: " + tagStr)
	}
	t.Name = cr[0]
	t.Operator = cr[1]
	t.Value = cr[2]
	return nil
}

type ValueQuery struct {
	Channel     *Channel
	Field       string
	Tags        []*Tag
	SummaryType string
	TimeStart   time.Time
	TimeEnd     time.Time
}

func (q *ValueQuery) Parse(params map[string]string) error {
	if q.Field, found = params["field"]; !found {
		return errors.New("missing field")
	}
	if _, found := q.Channel.Fields[q.Field]; !found {
		return errors.New("undefined field in channel: " + q.Field)
	}

	if q.SummaryType, found = params["summary_type"]; !found {
		return errors.New("missing summary_type")
	} else if !StringSliceContains(SupportedSummaryTypes, q.SummaryType) {
		return errors.New("unsupported summary_type: " + q.SummaryType)
	}

	if timeRange, found = params["time_range"]; !found {
		return errors.New("missing time_range")
	} else {
		ranges := strings.Split(timeRange, ":")
		if len(ranges) != 2 || len(ranges[0]) == 0 {
			return errors.New("invalid time_range format: " + timeRange)
		}
		start, err := strconv.ParseInt(ranges[0], 10, 64)
		if err != nil {
			return errors.New("error parsing time_range: " + timeRange)
		}
		q.TimeStart = time.Unix(start, 0)

		if len(ranges[1]) > 0 {
			end, err := strconv.ParseInt(ranges[1], 10, 64)
			if err != nil {
				return errors.New("error parsing time_range: " + timeRange)
			}
			q.TimeEnd = time.Unix(end, 0)
		}
	}

	q.Tags = []*Tag{}
	if tagStr, found = params["tags"]; found {
		tags := strings.Split(tagStr, ",")
		for _, tag := range tags {
			t := &Tag{}
			err := t.ParseTag(tag)
			if err != nil {
				return err
			} else if !StringSliceContains(q.Channel.Tags, t.Name) {
				return errors.New("undefined tag in channel, tag: " + t.Name)
			} else {
				q.Tags = append(q.Tags, t)
			}
		}
	}

	return nil
}

func (q *ValueQuery) ESQuery() {

}
