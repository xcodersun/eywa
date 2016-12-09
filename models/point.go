package models

import (
	"encoding/json"
	"errors"
	"fmt"
	. "github.com/eywa/connections"
	. "github.com/eywa/utils"
	"net/url"
	"strconv"
	"time"
)

var jsonParsingErr = errors.New("json parsing err")
var urlParsingErr = errors.New("url parsing err")
var IndexTypeMessages = "messages"
var IndexTypeActivities = "activities"

type Point struct {
	ch   *Channel
	conn Connection
	msg  Message

	Id        string
	Timestamp time.Time
	Tags      map[string]string
	Fields    map[string]interface{}
}

// extra information about a point is mixed in here
func (p *Point) MarshalJSON() ([]byte, error) {
	j := make(map[string]interface{})
	j["device_id"] = p.conn.Identifier()
	j["channel_name"] = p.ch.Name
	j["timestamp"] = NanoToMilli(p.Timestamp.UnixNano())

	if p.msg.Type() == TypeDisconnectMessage {
		j["activity"] = p.msg.TypeString()

		if p.conn.ClosedAt().IsZero() {
			j["duration"] = NanoToMilli(time.Now().Sub(p.conn.CreatedAt()).Nanoseconds())
		} else {
			j["duration"] = NanoToMilli(p.conn.ClosedAt().Sub(p.conn.CreatedAt()).Nanoseconds())
		}
	} else if p.msg.Type() == TypeConnectMessage {
		j["activity"] = p.msg.TypeString()
	} else {
		j["message_type"] = p.msg.TypeString()
	}

	for k, v := range p.Tags {
		j[k] = v
	}

	for k, v := range p.Fields {
		j[k] = v
	}

	j["connection_type"] = p.conn.ConnectionType()

	return json.Marshal(j)
}

func (p *Point) parseJson() error {
	jsonValues := make(map[string]json.RawMessage)
	err := json.Unmarshal(p.msg.Payload(), &jsonValues)
	if err != nil {
		return jsonParsingErr
	}

	if _, found := jsonValues["timestamp"]; found {
		var timestamp int64
		err = json.Unmarshal(jsonValues["timestamp"], &timestamp)
		if err != nil {
			return err
		}
		sec := MilliSecToSec(timestamp)
		nano := MilliSecToNano(timestamp)
		p.Timestamp = time.Unix(sec, nano)
	} else {
		p.Timestamp = time.Now()
	}

	p.Tags = make(map[string]string)
	for _, tag := range p.ch.Tags {
		if tagV, found := jsonValues[tag]; found {
			var str string
			err := json.Unmarshal(tagV, &str)
			if err != nil {
				return err
			}
			p.Tags[tag] = str
		}
	}

	p.Fields = make(map[string]interface{})
	for fieldName, fieldType := range p.ch.Fields {
		if fieldValue, found := jsonValues[fieldName]; found {
			switch fieldType {
			case "int":
				var v int64
				err := json.Unmarshal(fieldValue, &v)
				if err != nil {
					return err
				}
				p.Fields[fieldName] = v
			case "float":
				var v float64
				err := json.Unmarshal(fieldValue, &v)
				if err != nil {
					return err
				}
				p.Fields[fieldName] = v
			case "boolean":
				var v bool
				err := json.Unmarshal(fieldValue, &v)
				if err != nil {
					return err
				}
				p.Fields[fieldName] = v
			}
		}
	}

	return nil
}

func (p *Point) parseUrl() error {
	urlValues, err := url.ParseQuery(string(p.msg.Payload()))
	if err != nil {
		return urlParsingErr
	}

	if ts := urlValues.Get("timestamp"); len(ts) > 0 {
		timestamp, err := strconv.ParseInt(ts, 10, 64)
		if err != nil {
			return err
		}
		sec := MilliSecToSec(timestamp)
		nano := MilliSecToNano(timestamp)
		p.Timestamp = time.Unix(sec, nano)
	} else {
		p.Timestamp = time.Now()
	}

	p.Tags = make(map[string]string)
	for _, tag := range p.ch.Tags {
		if tagV := urlValues.Get(tag); len(tagV) > 0 {
			p.Tags[tag] = tagV
		}
	}

	p.Fields = make(map[string]interface{})
	for fieldName, fieldType := range p.ch.Fields {
		if fieldValue := urlValues.Get(fieldName); len(fieldValue) > 0 {
			switch fieldType {
			case "int":
				v, err := strconv.ParseInt(fieldValue, 10, 64)
				if err != nil {
					return err
				}
				p.Fields[fieldName] = v
			case "float":
				v, err := strconv.ParseFloat(fieldValue, 64)
				if err != nil {
					return err
				}
				p.Fields[fieldName] = v
			case "boolean":
				if fieldValue == "true" {
					p.Fields[fieldName] = true
				} else if fieldValue == "false" {
					p.Fields[fieldName] = false
				} else {
					return errors.New("invalid boolean value: " + fieldValue)
				}
			}
		}
	}

	return nil
}

func (p *Point) Metadata(meta map[string]string) {
	for k, v := range meta {
		if StringSliceContains(InternalTags, k) || StringSliceContains(p.ch.Tags, k) {
			if _, found := p.Tags[k]; !found {
				p.Tags[k] = v
			}
		}
	}
}

func (p *Point) IndexType() string {
	if p.msg.Type() == TypeConnectMessage || p.msg.Type() == TypeDisconnectMessage {
		return IndexTypeActivities
	}
	return IndexTypeMessages
}

func NewPoint(id string, ch *Channel, conn Connection, m Message) (*Point, error) {
	p := &Point{
		ch:   ch,
		conn: conn,
		msg:  m,
		Id:   id,
	}

	err := p.parseJson()
	if err != nil && err == jsonParsingErr {
		err = p.parseUrl()
	}
	if err != nil {
		return nil, err
	}

	p.Metadata(conn.Metadata())
	return p, nil
}

func TimedIndexName(ch *Channel, ts time.Time) string {
	year, week := ts.ISOWeek()
	return fmt.Sprintf("channels.%d.%d-%d", ch.Id, year, week)
}
