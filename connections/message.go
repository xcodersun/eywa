package connections

import (
	"fmt"
	"strconv"
	"strings"
)

const (
	AsyncRequestMessage = 0
	SyncRequestMessage  = 1
	ResponseMessage     = 2
	CloseMessage        = 8
)

type Message struct {
	MessageType int
	MessageId   string
	Payload     string
}

func (m *Message) String() string {
	return fmt.Sprintf("%d|%s|%s", m.MessageType, m.MessageId, m.Payload)
}

func (m *Message) Marshal() string {
	return m.String()
}

func (m *Message) Valid() bool {
	return (m.MessageType == AsyncRequestMessage ||
		m.MessageType == SyncRequestMessage ||
		m.MessageType == ResponseMessage) && len(m.MessageId) > 0
}

func Marshal(m *Message) string {
	return m.Marshal()
}

func Unmarshal(raw string) (*Message, error) {
	fields := strings.SplitN(raw, "|", 3)
	if len(fields) != 3 {
		return nil, &MessageParsingError{
			message: fmt.Sprintf("expected 3 fields instead of %d, [%s]", len(fields), raw),
		}
	}

	msgType, err := strconv.Atoi(fields[0])
	if err != nil || (msgType != AsyncRequestMessage &&
		msgType != SyncRequestMessage && msgType != ResponseMessage) {
		return nil, &MessageParsingError{
			message: fmt.Sprintf("invalid messagetype, [%s]", raw),
		}
	}

	m := &Message{
		MessageType: msgType,
		MessageId:   fields[1],
		Payload:     fields[2],
	}
	if len(m.MessageId) == 0 {
		return nil, &MessageParsingError{
			message: fmt.Sprintf("empty messageid, [%s]", raw),
		}
	}

	return m, nil
}
