package connections

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

const (
	AsyncRequestMessage = 1
	SyncRequestMessage  = 2
	ResponseMessage     = 4
	CloseMessage        = 8
)

type Message struct {
	MessageType int
	MessageId   string
	Payload     string
}

type MessageResp struct {
	msg *Message
	err error
}

type MessageReq struct {
	msg    *Message
	respCh chan *MessageResp
}

func (m *Message) String() string {
	return fmt.Sprintf("%d|%s|%s", m.MessageType, m.MessageId, m.Payload)
}

func (m *Message) Marshal() string {
	return m.String()
}

func Marshal(m *Message) string {
	return m.Marshal()
}

func Unmarshal(raw string) (*Message, error) {
	fields := strings.SplitN(raw, "|", 3)
	if len(fields) != 3 {
		return nil, errors.New(fmt.Sprintf("expected 3 fields instead of %d, raw: %s in message", len(fields), raw))
	}

	msgType, err := strconv.Atoi(fields[0])
	if err != nil || (msgType != AsyncRequestMessage &&
		msgType != SyncRequestMessage && msgType != ResponseMessage && msgType != CloseMessage) {
		return nil, errors.New(fmt.Sprintf("invalid MessageType, raw: %s", raw))
	}

	m := &Message{
		MessageType: msgType,
		MessageId:   fields[1],
		Payload:     fields[2],
	}
	if len(m.MessageId) == 0 && msgType != CloseMessage {
		return nil, errors.New("empty MessageId")
	}

	return m, nil
}
