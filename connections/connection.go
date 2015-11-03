package connections

import (
	"fmt"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/spf13/viper"
)

func NewConnection(identifier string, ws *websocket.Conn, h MessageHandler) (Connection, error) {
	conn := &connection{
		identifier: identifier,
		createdAt:  time.Now(),
		closed:     false,
		msgChans:   make(map[string]chan *Message),
		ws:         ws,
	}
	ws.SetPingHandler(func(payload string) error {
		return CM.RegisterConnection(conn)
	})

	if err := CM.RegisterConnection(conn); err != nil {
		return nil, err
	}

	go conn.Listen(h)
	return conn, nil
}

type Connection interface {
	Identifier() string
	Closed() bool
	CreatedAt() time.Time

	Close() error
	SendAsyncRequest(*Message) error
	SendSyncRequest(*Message) (*Message, error)
	SendResponse(*Message) error
	Listen(MessageHandler)
}

type connection struct {
	createdAt  time.Time
	identifier string

	closed   bool
	msgChans map[string]chan *Message
	ws       *websocket.Conn

	rm  sync.Mutex
	wm  sync.Mutex
	chm sync.RWMutex
	wg  sync.WaitGroup
}

func (c *connection) Identifier() string {
	return c.identifier
}

func (c *connection) CreatedAt() time.Time {
	return c.createdAt
}

func (c *connection) Closed() bool {
	return c.closed
}

func (c *connection) sendMessage(message *Message) error {
	c.wm.Lock()
	defer c.wm.Unlock()

	if c.closed {
		return &MessageSendingError{message: "connection closed"}
	}

	err := c.ws.SetWriteDeadline(time.Now().Add(viper.GetDuration("connections.timeouts.write")))
	if err != nil {
		return &MessageSendingError{message: "error setting write deadline, " + err.Error()}
	}

	if message.MessageType == CloseMessage {
		err = c.ws.WriteMessage(websocket.CloseMessage, []byte(message.Payload))
	} else {
		err = c.ws.WriteMessage(websocket.TextMessage, []byte(message.Marshal()))
	}

	if err != nil {
		err = &MessageSendingError{message: err.Error()}
	}

	return err
}

func (c *connection) SendAsyncRequest(message *Message) error {
	if message.MessageType != AsyncRequestMessage {
		return &MessageSendingError{
			message: fmt.Sprintf("invalid message type %s, expected %s", message.MessageType, AsyncRequestMessage),
		}
	}

	return c.sendMessage(message)
}

func (c *connection) SendResponse(message *Message) error {
	if message.MessageType != ResponseMessage {
		return &MessageSendingError{
			message: fmt.Sprintf("invalid message type %s, expected %s", message.MessageType, ResponseMessage),
		}
	}

	return c.sendMessage(message)
}

func (c *connection) SendSyncRequest(message *Message) (*Message, error) {
	if message.MessageType != SyncRequestMessage {
		return nil, &MessageSendingError{
			message: fmt.Sprintf("invalid message type %s, expected %s", message.MessageType, SyncRequestMessage),
		}
	}

	msgChan := make(chan *Message, 1)
	c.chm.Lock()
	c.msgChans[message.MessageId] = msgChan
	c.chm.Unlock()
	c.wg.Add(1)
	defer func() {
		c.chm.Lock()
		delete(c.msgChans, message.MessageId)
		c.chm.Unlock()
		c.wg.Done()
	}()

	err := c.sendMessage(message)
	if err != nil {
		return nil, err
	}

	select {
	case <-time.After(viper.GetDuration("connections.timeouts.response")):
		return nil, &ResponseTimeoutError{
			message: fmt.Sprintf("response timed out for %s", viper.GetDuration("connections.timeouts.response")),
		}
	case resp := <-msgChan:
		return resp, nil
	}
}

func (c *connection) lock() {
	c.wm.Lock()
	c.rm.Lock()
}

func (c *connection) unlock() {
	c.rm.Unlock()
	c.wm.Unlock()
}

func (c *connection) Close() error {
	c.lock()

	if c.closed {
		c.unlock()
		return nil
	}

	c.closed = true
	if err := c.ws.SetWriteDeadline(time.Now().Add(viper.GetDuration("connections.timeouts.write"))); err != nil {
		c.unlock()
		return &ConnectionCloseError{
			message: fmt.Sprintf("error setting write deadline, %s", err.Error()),
		}
	}
	if err := c.ws.WriteMessage(websocket.CloseMessage, []byte{}); err != nil {
		c.unlock()
		return &ConnectionCloseError{
			message: fmt.Sprintf("error writing close control, %s", err.Error()),
		}
	}
	if err := c.ws.Close(); err != nil {
		c.unlock()
		return &ConnectionCloseError{
			message: fmt.Sprintf("error closing websocket, %s", err.Error()),
		}
	}

	c.unlock()

	c.wg.Wait()
	if err := CM.UnregisterConnection(c); err != nil {
		return &ConnectionUnregisterError{
			message: fmt.Sprintf("error unregistering connection, %s", err.Error()),
		}
	}

	return nil
}

func (c *connection) readMessage() (*Message, error) {
	c.rm.Lock()
	defer c.rm.Unlock()

	if c.closed {
		return nil, &MessageReadingError{message: "connection is closed"}
	}

	if err := c.ws.SetReadDeadline(time.Now().Add(viper.GetDuration("connections.timeouts.read"))); err != nil {
		return nil, &MessageReadingError{
			message: fmt.Sprintf("error setting read deadline, %s", err.Error()),
		}
	}

	messageType, messageBody, err := c.ws.ReadMessage()
	if err != nil {
		return nil, &MessageReadingError{
			message: fmt.Sprintf("error reading message, %s", err.Error()),
		}
	}

	if messageType == websocket.CloseMessage {
		return &Message{
			MessageType: CloseMessage,
		}, nil
	}

	return Unmarshal(string(messageBody))
}

func (c *connection) Listen(h MessageHandler) {
	for {
		message, err := c.readMessage()
		if err != nil {
			c.Close()
			h(c, nil, err)
			return
		} else if message.MessageType == CloseMessage {
			c.Close()
			h(c, message, nil)
			return
		} else if message.MessageType == ResponseMessage {
			msgId := message.MessageId
			c.chm.RLock()
			ch, found := c.msgChans[msgId]
			c.chm.RUnlock()
			if found {
				ch <- message
				DefaultMiddlewares.Chain(nil)(c, message, nil)
			} else {
				h(c, nil, &MessageResponseError{
					message: "unexpected response messages received, probably due to connection reset?",
				})
			}
		} else {
			h(c, message, nil)
		}
	}
}
