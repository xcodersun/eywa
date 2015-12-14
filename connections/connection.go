package connections

import (
	"errors"
	"fmt"
	"github.com/gorilla/websocket"
	. "github.com/vivowares/octopus/configs"
	"io"
	"net"
	"strconv"
	"sync"
	"time"
)

type WebsocketError struct {
	message string
}

func (e *WebsocketError) Error() string {
	return fmt.Sprintf("WebsocketError: %s", e.message)
}

type syncRespChanMap struct {
	sync.Mutex
	m map[string]chan *MessageResp
}

func (sm *syncRespChanMap) put(msgId string, ch chan *MessageResp) {
	sm.Lock()
	defer sm.Unlock()

	sm.m[msgId] = ch
}

func (sm *syncRespChanMap) find(msgId string) (chan *MessageResp, bool) {
	sm.Lock()
	defer sm.Unlock()

	ch, found := sm.m[msgId]
	return ch, found
}

func (sm *syncRespChanMap) delete(msgId string) {
	sm.Lock()
	defer sm.Unlock()

	delete(sm.m, msgId)
}

func (sm *syncRespChanMap) len() int {
	sm.Lock()
	defer sm.Unlock()

	return len(sm.m)
}

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

type Connection struct {
	shard        *shard
	ws           wsConn
	createdAt    time.Time
	lastPingedAt time.Time
	identifier   string
	h            MessageHandler
	Metadata     map[string]interface{}

	wg        sync.WaitGroup
	closeOnce sync.Once
	closed    bool

	// there is a chance for this msgChan to grow,
	// in extreme race condition. no plan to fix it.
	// simple solution is to limit the size of it,
	// close the connection when it blows up.
	msgChans *syncRespChanMap

	wch      chan *MessageReq // size=?
	closewch chan bool        // size=1
	rch      chan struct{}    // size=0
}

func (c *Connection) Identifier() string { return c.identifier }

func (c *Connection) CreatedAt() time.Time { return c.createdAt }

func (c *Connection) LastPingedAt() time.Time { return c.lastPingedAt }

func (c *Connection) Closed() bool { return c.closed }

func (c *Connection) SendAsyncRequest(msg string) error {
	_, err := c.sendMessage(AsyncRequestMessage, msg)
	return err
}

func (c *Connection) SendResponse(msg string) error {
	_, err := c.sendMessage(ResponseMessage, msg)
	return err
}

func (c *Connection) SendSyncRequest(msg string) (string, error) {
	return c.sendMessage(SyncRequestMessage, msg)
}

func (c *Connection) sendMessage(messageType int, payload string) (respMsg string, err error) {
	respMsg = ""

	defer func() {
		if r := recover(); r != nil {
			err = errors.New("connection is closed")
		}
	}()

	msgId := strconv.FormatInt(time.Now().UnixNano(), 16)
	msg := &Message{
		MessageType: messageType,
		MessageId:   msgId,
		Payload:     payload,
	}

	respCh := make(chan *MessageResp, 1)

	timeout := Config.Connections.Timeouts.Request
	select {
	case <-time.After(timeout):
		err = errors.New(fmt.Sprintf("request timed out for %s", timeout))
		return
	case c.wch <- &MessageReq{
		msg:    msg,
		respCh: respCh,
	}:
	}

	if messageType == SyncRequestMessage {
		defer func() {
			c.msgChans.delete(msgId)
		}()

		timeout = Config.Connections.Timeouts.Response
	}

	select {
	case <-time.After(timeout):
		err = errors.New(fmt.Sprintf("response timed out for %s", timeout))
		return
	case resp := <-respCh:
		if resp.msg != nil {
			respMsg = resp.msg.Payload
		}
		err = resp.err
		return
	}

}

func (c *Connection) wListen() {
	defer c.wg.Done()
	for {
		req, more := <-c.wch
		if more {
			err := c.sendWsMessage(req.msg)
			if err != nil {
				req.respCh <- &MessageResp{
					msg: nil,
					err: err,
				}

				if _, ok := err.(*WebsocketError); ok {
					c.Close()
				}
			} else {
				if req.msg.MessageType == SyncRequestMessage {
					c.msgChans.put(req.msg.MessageId, req.respCh)
				} else {
					req.respCh <- &MessageResp{}
				}
			}
		} else {
			<-c.closewch
			c.sendWsMessage(&Message{MessageType: CloseMessage})
			return
		}
	}
}

func (c *Connection) sendWsMessage(message *Message) error {
	err := c.ws.SetWriteDeadline(time.Now().Add(Config.Connections.Timeouts.Write))
	if err != nil {
		return &WebsocketError{message: "error setting write deadline, " + err.Error()}
	}

	if message.MessageType == CloseMessage {
		err = c.ws.WriteMessage(websocket.CloseMessage, []byte(message.Payload))
		err = c.ws.Close()
	} else {
		err = c.ws.WriteMessage(websocket.TextMessage, []byte(message.Marshal()))
	}

	if err != nil {
		err = &WebsocketError{message: err.Error()}
	}

	return err
}

func (c *Connection) readWsMessage() (*Message, error) {
	if err := c.ws.SetReadDeadline(time.Now().Add(Config.Connections.Timeouts.Read)); err != nil {
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
	c.shard.updateRegistry(c)

	if messageType == websocket.CloseMessage {
		return &Message{
			MessageType: CloseMessage,
		}, nil
	}

	return Unmarshal(string(messageBody))
}

func (c *Connection) rListen() {
	defer c.wg.Done()
	for {
		select {
		case <-c.rch:
			return
		default:
			message, err := c.readWsMessage()
			if err != nil {
				c.h(c, nil, err)
				if _, ok := err.(*WebsocketError); ok {
					c.Close()
					return
				}
			} else if message.MessageType == CloseMessage {
				c.h(c, message, nil)
				c.Close()
				return
			} else if message.MessageType == ResponseMessage {
				ch, found := c.msgChans.find(message.MessageId)
				if found {
					c.msgChans.delete(message.MessageId)
					ch <- &MessageResp{msg: message}
					DefaultMiddlewares.Chain(nil)(c, message, nil)
				} else {
					c.h(c, message, errors.New("unexpected response messages received, probably due to response timeout?"))
				}
			} else {
				c.h(c, message, nil)
			}
		}
	}
}

func (c *Connection) Close() {
	c.closeOnce.Do(func() {
		c.closed = true
		close(c.wch)
		close(c.rch)
		c.closewch <- true
		c.shard.unregister(c)
	})
}

func (c *Connection) Wait() {
	c.wg.Wait()
}

func (c *Connection) Start() {
	c.wg.Add(2)
	go c.rListen()
	go c.wListen()
}
