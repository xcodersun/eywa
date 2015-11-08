package connections

import (
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/spf13/viper"
)

type wsConn interface {
	Subprotocol() string
	Close() error
	LocalAddr() net.Addr
	RemoteAddr() net.Addr
	WriteControl(int, []byte, time.Time) error
	NextWriter(int) (io.WriteCloser, error)
	WriteMessage(int, []byte) error
	SetWriteDeadline(time.Time) error
	NextReader() (int, io.Reader, error)
	ReadMessage() (int, []byte, error)
	SetReadDeadline(time.Time) error
	SetReadLimit(int64)
	SetPingHandler(h func(string) error)
	SetPongHandler(h func(string) error)
	UnderlyingConn() net.Conn
}

type Connection interface {
	Host() string
	IsLocal() bool
	Identifier() string
	Closed() bool
	CreatedAt() time.Time
	LastPingedAt() time.Time

	SendAsyncRequest(*Message) error
	SendSyncRequest(*Message) (*Message, error)
	SendResponse(*Message) error

	listen(MessageHandler)
	close() error
	signalClose()
}

type connection struct {
	createdAt    time.Time
	lastPingedAt time.Time
	identifier   string
	cm           ConnectionManager
	closeChan    chan bool
	closed       bool
	msgChans     map[string]chan *Message
	ws           wsConn

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

func (c *connection) LastPingedAt() time.Time {
	c.rm.Lock()
	defer c.rm.Unlock()
	return c.lastPingedAt
}

func (c *connection) Host() string {
	return c.cm.Host()
}

func (c *connection) IsLocal() bool {
	return c.cm.Host() == viper.GetString("host")
}

func (c *connection) Closed() bool {
	return c.closed
}

func (c *connection) signalClose() {
	c.closeChan <- true
}

func (c *connection) readMessage() (*Message, error) {
	c.rm.Lock()
	defer c.rm.Unlock()

	if c.closed {
		return nil, &MessageReadingError{message: "connection is closed"}
	}

	if err := c.ws.SetReadDeadline(time.Now().Add(viper.GetDuration("connections.timeouts.read"))); err != nil {
		return nil, &WebsocketError{
			message: fmt.Sprintf("error setting read deadline, %s", err.Error()),
		}
	}

	messageType, messageBody, err := c.ws.ReadMessage()
	if err != nil {
		return nil, &WebsocketError{
			message: fmt.Sprintf("error reading message, %s", err.Error()),
		}
	}

	c.lastPingedAt = time.Now()
	c.cm.refreshConnectionRegistry(c, c.lastPingedAt)

	if messageType == websocket.CloseMessage {
		return &Message{
			MessageType: CloseMessage,
		}, nil
	}

	return Unmarshal(string(messageBody))
}

func (c *connection) sendMessage(message *Message) error {
	if len(message.MessageId) == 0 {
		return &MessageIdError{
			message: "empty message id",
		}
	}

	c.wm.Lock()
	defer c.wm.Unlock()

	if c.closed {
		return &MessageSendingError{message: "connection closed"}
	}

	err := c.ws.SetWriteDeadline(time.Now().Add(viper.GetDuration("connections.timeouts.write")))
	if err != nil {
		return &WebsocketError{message: "error setting write deadline, " + err.Error()}
	}

	if message.MessageType == CloseMessage {
		err = c.ws.WriteMessage(websocket.CloseMessage, []byte(message.Payload))
	} else {
		err = c.ws.WriteMessage(websocket.TextMessage, []byte(message.Marshal()))
	}

	if err != nil {
		err = &WebsocketError{message: err.Error()}
	}

	return err
}

func (c *connection) SendAsyncRequest(message *Message) error {
	if message.MessageType != AsyncRequestMessage {
		return &MessageSendingError{
			message: fmt.Sprintf("invalid message type %d, expected %d", message.MessageType, AsyncRequestMessage),
		}
	}

	return c.sendMessage(message)
}

func (c *connection) SendResponse(message *Message) error {
	if message.MessageType != ResponseMessage {
		return &MessageSendingError{
			message: fmt.Sprintf("invalid message type %d, expected %d", message.MessageType, ResponseMessage),
		}
	}

	return c.sendMessage(message)
}

func (c *connection) createMessageChan(messageId string) chan *Message {
	msgChan := make(chan *Message, 1)
	c.chm.Lock()
	c.msgChans[messageId] = msgChan
	c.chm.Unlock()
	c.wg.Add(1)
	return msgChan
}

func (c *connection) removeMessageChan(messageId string) {
	c.chm.Lock()
	delete(c.msgChans, messageId)
	c.chm.Unlock()
	c.wg.Done()
}

func (c *connection) findMessageChan(messageId string) (chan *Message, bool) {
	c.chm.RLock()
	defer c.chm.RUnlock()
	m, found := c.msgChans[messageId]
	return m, found
}

func (c *connection) SendSyncRequest(message *Message) (*Message, error) {
	if message.MessageType != SyncRequestMessage {
		return nil, &MessageSendingError{
			message: fmt.Sprintf("invalid message type %d, expected %d", message.MessageType, SyncRequestMessage),
		}
	}

	ch := c.createMessageChan(message.MessageId)
	defer c.removeMessageChan(message.MessageId)

	err := c.sendMessage(message)
	if err != nil {
		return nil, err
	}

	select {
	case <-time.After(viper.GetDuration("connections.timeouts.response")):
		return nil, &ResponseTimeoutError{
			message: fmt.Sprintf("response timed out for %s", viper.GetDuration("connections.timeouts.response")),
		}
	case resp := <-ch:
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

func (c *connection) close() (err error) {
	err = nil

	c.lock()

	if c.closed {
		c.unlock()
		return
	}

	c.closed = true
	if err = c.ws.SetWriteDeadline(time.Now().Add(viper.GetDuration("connections.timeouts.write"))); err != nil {
		err = &ConnectionCloseError{
			message: fmt.Sprintf("error setting write deadline, %s", err.Error()),
		}
	} else {
		if err = c.ws.WriteMessage(websocket.CloseMessage, []byte{}); err != nil {
			err = &ConnectionCloseError{
				message: fmt.Sprintf("error writing close control, %s", err.Error()),
			}
		}
	}

	if err = c.ws.Close(); err != nil {
		err = &ConnectionCloseError{
			message: fmt.Sprintf("error closing websocket, %s", err.Error()),
		}
	}
	c.unlock()

	c.cm.unregisterConnection(c)

	c.wg.Wait()

	return
}

func (c *connection) listen(h MessageHandler) {
	for {
		select {
		case <-c.closeChan:
			c.close()
			return
		default:
			message, err := c.readMessage()
			if err != nil {
				h(c, nil, err)
				if _, ok := err.(*WebsocketError); ok {
					c.close()
					return
				}
			} else if message.MessageType == CloseMessage {
				h(c, message, nil)
				c.close()
				return
			} else if message.MessageType == ResponseMessage {
				ch, found := c.findMessageChan(message.MessageId)
				if found {
					ch <- message
					DefaultMiddlewares.Chain(nil)(c, message, nil)
				} else {
					h(c, message, &MessageResponseError{
						message: "unexpected response messages received, probably due to connection reset?",
					})
				}
			} else {
				h(c, message, nil)
			}
		}
	}
}
