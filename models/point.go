package models

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/vivowares/octopus/connections"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// use point to write
// point data is stored in influxdb or elasticsearch
// we also only support unix time epoch as timestamp
type Point struct {
	channel *Channel // goes to archive with name
	conn    connections.Connection
	Raw     string // goes to archive

	Timestamp time.Time
	Tags      map[string]string
	Fields    map[string]interface{}
}

func (p *Point) Store() error {
	return IStore.WritePoint(p)
}

type stringGetter interface {
	Get(string) string
}

type mapGetter map[string]interface{}

func (m mapGetter) Get(key string) string {
	if v, found := m[key]; found {
		switch v.(type) {
		case int64:
			return fmt.Sprintf("%d", v)
		case float64:
			return fmt.Sprintf("%f", v)
		default:
			return fmt.Sprintf("%v", v)
		}
	}
	return ""
}

func (p *Point) parseRaw() error {
	var sg stringGetter
	switch p.channel.Format {
	case "url":
		urlValues, err := url.ParseQuery(p.Raw)
		if err != nil {
			return err
		}
		sg = urlValues
	case "json":
		jsonValues := make(map[string]interface{})
		d := json.NewDecoder(strings.NewReader(p.Raw))
		d.UseNumber()
		err := d.Decode(&jsonValues)
		if err != nil {
			return err
		}
		sg = mapGetter(jsonValues)
	default:
		return errors.New("unsupported point format, supported formats are: " + strings.Join(SupportedPointFormat, ","))
	}

	if len(sg.Get("timestamp")) > 0 {
		t, err := strconv.ParseInt(sg.Get("timestamp"), 10, 64)
		if err != nil {
			return err
		} else {
			nano := int64(time.Now().Nanosecond()) % int64(time.Millisecond)
			p.Timestamp = time.Unix(0, t*int64(time.Millisecond)+nano)
		}
	} else {
		return errors.New("missing timestamp in point")
	}

	tags := make(map[string]string)
	for _, tag := range p.channel.Tags {
		if len(sg.Get(tag)) > 0 {
			tags[tag] = sg.Get(tag)
		}
	}
	tags["connection_id"] = p.conn.Identifier()
	p.Tags = tags

	fields := make(map[string]interface{})
	for fieldName, fieldType := range p.channel.Fields {
		fieldValue := sg.Get(fieldName)
		if len(fieldValue) > 0 {
			switch fieldType {
			case "boolean":
				if fieldValue == "true" {
					fields[fieldName] = true
				} else if fieldValue == "false" {
					fields[fieldName] = false
				} else {
					return errors.New("invalid boolean value: " + fieldValue)
				}
			case "int":
				i, err := strconv.ParseInt(fieldValue, 10, 64)
				if err != nil {
					return err
				}
				fields[fieldName] = i
			case "float":
				f, err := strconv.ParseFloat(fieldValue, 64)
				if err != nil {
					return err
				}
				fields[fieldName] = f
			default:
				fields[fieldName] = fieldValue
			}
		}
	}
	p.Fields = fields

	return nil
}
