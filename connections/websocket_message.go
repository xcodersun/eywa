package connections

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"
	"time"
)

var SupportedWebsocketMessageTypes = map[MessageType]string{
	TypeUploadMessage:     "upload",
	TypeResponseMessage:   "response",
	TypeSendMessage:       "send",
	TypeRequestMessage:    "request",
	TypeConnectMessage:    "connect",
	TypeDisconnectMessage: "disconnect",
}

type websocketMessageResp struct {
	msg *websocketMessage
	err error
}

type websocketMessageReq struct {
	msg    *websocketMessage
	respCh chan *websocketMessageResp
}

type websocketMessage struct {
	_type   MessageType
	id      string
	payload []byte
	raw     []byte
}

func NewWebsocketMessage(t MessageType, id string, payload []byte, raw []byte) *websocketMessage {
	return &websocketMessage{_type: t, id: id, payload: payload, raw: raw}
}

func (m *websocketMessage) TypeString() string { return SupportedWebsocketMessageTypes[m._type] }
func (m *websocketMessage) Type() MessageType  { return m._type }
func (m *websocketMessage) Id() string         { return m.id }
func (m *websocketMessage) Payload() []byte    { return m.payload }
func (m *websocketMessage) Raw() []byte        { return m.raw }

func (m *websocketMessage) Marshal() ([]byte, error) {
	if _, found := SupportedWebsocketMessageTypes[m._type]; !found {
		return nil, errors.New(fmt.Sprintf("unsupported websocket message type %d", m._type))
	}

	if m.payload == nil {
		if m._type != TypeDisconnectMessage && m._type != TypeConnectMessage {
			return nil, errors.New(fmt.Sprintf("missing payload for websocket message type %s", SupportedWebsocketMessageTypes[m._type]))
		} else {
			m.payload = []byte{}
		}
	}

	if len(m.id) == 0 {
		if m._type == TypeRequestMessage || m._type == TypeResponseMessage {
			return nil, errors.New(fmt.Sprintf("missing message id for websocket message type %s", SupportedWebsocketMessageTypes[m._type]))
		} else {
			m.id = strconv.FormatInt(time.Now().UnixNano(), 16)
		}
	}

	p := bytes.Buffer{}

	_, err := p.WriteString(fmt.Sprintf("%d|%s|", m._type, m.id))
	if err != nil {
		return nil, errors.New(fmt.Sprintf("error formatting message, %s", err.Error()))
	}

	_, err = p.Write(m.payload)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("error formatting message, %s", err.Error()))
	}

	m.raw = p.Bytes()
	return m.raw, nil
}

func (m *websocketMessage) Unmarshal() error {
	if m.raw == nil {
		if m._type == TypeConnectMessage || m._type == TypeDisconnectMessage {
			m.raw = []byte{}
			m.payload = m.raw
			if len(m.id) == 0 {
				m.id = strconv.FormatInt(time.Now().UnixNano(), 16)
			}
			return nil
		} else {
			return errors.New("missing message body for websocket message")
		}
	}

	pips := 0
	var pip1, pip2 int
	for idx, b := range m.raw {
		if b == '|' {
			pips += 1
			if pips == 1 {
				pip1 = idx
				msgType, err := strconv.Atoi(string(m.raw[0:idx]))
				if err != nil {
					return errors.New(fmt.Sprintf("error parsing websocket message type, %s", err.Error()))
				} else {
					m._type = MessageType(msgType)
				}
			} else if pips == 2 {
				pip2 = idx
				m.id = string(m.raw[pip1+1 : pip2])
				m.payload = m.raw[pip2+1:]
				break
			} else {
				break
			}
		}
	}

	if _, found := SupportedWebsocketMessageTypes[m._type]; !found {
		return errors.New(fmt.Sprintf("unsupported websocket message type %d", m._type))
	}

	if pips != 2 && m._type != TypeDisconnectMessage && m._type != TypeConnectMessage {
		return errors.New(fmt.Sprintf("expected 2 pips instead of %d for websocket message type %s", pips, SupportedWebsocketMessageTypes[m._type]))
	}

	if len(m.id) == 0 {
		if m._type == TypeRequestMessage || m._type == TypeResponseMessage {
			return errors.New(fmt.Sprintf("empty message id for websocket message type %s", SupportedWebsocketMessageTypes[m._type]))
		} else {
			m.id = strconv.FormatInt(time.Now().UnixNano(), 16)
		}
	}

	return nil
}
