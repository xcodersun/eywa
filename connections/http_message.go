package connections

import (
	"errors"
	"fmt"
	"strconv"
	"time"
)

var SupportedHttpMessageTypes = map[MessageType]string{
	TypeUploadMessage:     "upload",
	TypeSendMessage:       "send",
	TypeConnectMessage:    "connect",
	TypeDisconnectMessage: "disconnect",
}

type httpMessage struct {
	_type MessageType
	id    string
	raw   []byte
}

func NewHttpMessage(t MessageType, id string, raw []byte) *httpMessage {
	return &httpMessage{_type: t, id: id, raw: raw}
}

func (m *httpMessage) TypeString() string { return SupportedHttpMessageTypes[m._type] }
func (m *httpMessage) Type() MessageType  { return m._type }
func (m *httpMessage) Id() string         { return m.id }
func (m *httpMessage) Payload() []byte    { return m.raw }
func (m *httpMessage) Raw() []byte        { return m.raw }

func (m *httpMessage) Marshal() ([]byte, error) {
	if _, found := SupportedHttpMessageTypes[m._type]; !found {
		return nil, errors.New(fmt.Sprintf("unsupported http message type %d", m._type))
	}

	if m.raw == nil {
		if m._type != TypeDisconnectMessage && m._type != TypeConnectMessage {
			return nil, errors.New(fmt.Sprintf("missing message body for http message type %s", SupportedHttpMessageTypes[m._type]))
		} else {
			m.raw = []byte{}
		}
	}

	if len(m.id) == 0 {
		m.id = strconv.FormatInt(time.Now().UnixNano(), 16)
	}

	return m.raw, nil
}

func (m *httpMessage) Unmarshal() error {
	if _, found := SupportedHttpMessageTypes[m._type]; !found {
		return errors.New(fmt.Sprintf("unsupported http message type %d", m._type))
	}

	if m.raw == nil {
		if m._type != TypeDisconnectMessage && m._type != TypeConnectMessage {
			return errors.New(fmt.Sprintf("missing message body for http message type %s", SupportedHttpMessageTypes[m._type]))
		} else {
			m.raw = []byte{}
		}
	}

	if len(m.id) == 0 {
		m.id = strconv.FormatInt(time.Now().UnixNano(), 16)
	}

	return nil
}
