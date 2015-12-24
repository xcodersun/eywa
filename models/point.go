package models

import (
	"encoding/json"
	"errors"
	. "github.com/vivowares/octopus/connections"
	"time"
)

type Point struct {
	ch   *Channel
	conn *Connection
	msg  *Message

	Id        string
	Timestamp time.Time
	Tags      map[string]string
	Fields    map[string]interface{}
}

func (p *Point) MarshalJSON() ([]byte, error) {
	j := make(map[string]interface{})
	j["id"] = p.Id
	j["connection_id"] = p.conn.Identifier()
	j["channel_name"] = p.ch.Name
	j["channel_id"] = p.ch.Id
	j["channel_based64_id"] = p.ch.Base64Id()
	j["timestamp"] = p.Timestamp.UTC().UnixNano() / int64(time.Millisecond)
	j["message_id"] = p.msg.MessageId
	switch p.msg.MessageType {
	case ResponseMessage:
		j["message_type"] = "response"
	case AsyncRequestMessage:
		j["message_type"] = "async_request"
	case SyncRequestMessage:
		j["message_type"] = "sync_request"
	case CloseMessage:
		j["message_type"] = "close"
	}

	for k, v := range p.Tags {
		j[k] = v
	}

	for k, v := range p.Fields {
		j[k] = v
	}

	return json.Marshal(j)
}

func (p *Point) parseJson() error {
	jsonValues := make(map[string]json.RawMessage)
	err := json.Unmarshal([]byte(p.msg.Payload), &jsonValues)
	if err != nil {
		return err
	}

	if _, found := jsonValues["timestamp"]; found {
		var timestamp int64
		err = json.Unmarshal(jsonValues["timestamp"], &timestamp)
		if err != nil {
			return err
		}
		p.Timestamp = time.Unix(timestamp, 0)
	} else {
		return errors.New("missing timestamp")
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
			case "string":
				var v string
				err := json.Unmarshal(fieldValue, &v)
				if err != nil {
					return err
				}
				p.Fields[fieldName] = v
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

func NewPoint(id string, ch *Channel, conn *Connection, m *Message) (*Point, error) {
	p := &Point{
		ch:   ch,
		conn: conn,
		msg:  m,
		Id:   id,
	}
	err := p.parseJson()
	if err != nil {
		return nil, err
	}
	return p, nil
}
