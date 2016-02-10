package connections

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"
)

const (
	AsyncRequestMessage = 1
	SyncRequestMessage  = 2
	ResponseMessage     = 4

	// these two messages are only used for connection states internally
	StartMessage = 0
	CloseMessage = 8
)

type Message struct {
	MessageType int
	MessageId   string
	Payload     []byte
}

type MessageResp struct {
	msg *Message
	err error
}

type MessageReq struct {
	msg    *Message
	respCh chan *MessageResp
}

func (m *Message) Marshal() ([]byte, error) {
	p := bytes.Buffer{}

	_, err := p.WriteString(fmt.Sprintf("%d|%s|", m.MessageType, m.MessageId))
	if err != nil {
		return nil, err
	}

	_, err = p.Write(m.Payload)
	if err != nil {
		return nil, err
	}

	return p.Bytes(), nil
}

func Marshal(m *Message) ([]byte, error) {
	return m.Marshal()
}

func Unmarshal(raw []byte) (*Message, error) {
	msg := &Message{}
	pips := 0
	var pip1, pip2 int
	for idx, b := range raw {
		if b == '|' {
			pips += 1
			if pips == 1 {
				pip1 = idx
				msgType, err := strconv.Atoi(string(raw[0:idx]))
				if err != nil {
					return nil, err
				} else if msgType != AsyncRequestMessage &&
					msgType != SyncRequestMessage && msgType != ResponseMessage && msgType != CloseMessage {
					return nil, errors.New(fmt.Sprintf("invalid MessageType %d, raw: %s", msgType, raw))
				} else {
					msg.MessageType = msgType
				}
			} else if pips == 2 {
				pip2 = idx
				msg.MessageId = string(raw[pip1+1 : pip2])
				break
			} else {
				break
			}
		}
	}
	if pips != 2 {
		return nil, errors.New(fmt.Sprintf("expected 2 pips instead of %d, raw: %s", pips, raw))
	}

	if len(msg.MessageId) == 0 && msg.MessageType != CloseMessage {
		return nil, errors.New("empty MessageId")
	}

	msg.Payload = raw[pip2+1:]
	return msg, nil
}
