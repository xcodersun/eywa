package connections

import (
	"errors"
	"fmt"
	"github.com/vivowares/eywa/Godeps/_workspace/src/github.com/gorilla/websocket"
	. "github.com/vivowares/eywa/configs"
	. "github.com/vivowares/eywa/utils"
	"io"
	"net"
	"strconv"
	"sync"
	"time"
)

var closedConnErr = errors.New("connection is closed")
var unexpectedMessageErr = errors.New("unexpected response messages received, probably due to response timeout?")

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

type Connection interface {
	Identifier() string
	Metadata() map[string]interface{}
	MessageHandler() MessageHandler
}

type HttpConnection struct {
	identifier string
	h          MessageHandler
	metadata   map[string]interface{}
}

func (c *HttpConnection) Identifier() string { return c.identifier }

func (c *HttpConnection) Metadata() map[string]interface{} { return c.metadata }

func (c *HttpConnection) MessageHandler() MessageHandler { return c.h }

type WebSocketConnection struct {
	shard        *shard
	ws           wsConn
	createdAt    time.Time
	lastPingedAt time.Time
	closedAt     time.Time
	identifier   string
	h            MessageHandler
	metadata     map[string]interface{}

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

func (c *WebSocketConnection) Identifier() string { return c.identifier }

func (c *WebSocketConnection) CreatedAt() time.Time { return c.createdAt }

func (c *WebSocketConnection) LastPingedAt() time.Time { return c.lastPingedAt }

func (c *WebSocketConnection) Closed() bool { return c.closed }

func (c *WebSocketConnection) MessageHandler() MessageHandler { return c.h }

func (c *WebSocketConnection) Metadata() map[string]interface{} { return c.metadata }

func (c *WebSocketConnection) Send(msg []byte) error {
	_, err := c.sendMessage(TypeSendMessage, msg)
	return err
}

func (c *WebSocketConnection) Response(msg []byte) error {
	_, err := c.sendMessage(TypeResponseMessage, msg)
	return err
}

func (c *WebSocketConnection) Request(msg []byte) ([]byte, error) {
	return c.sendMessage(TypeRequestMessage, msg)
}

func (c *WebSocketConnection) sendMessage(messageType int, payload []byte) (respMsg []byte, err error) {
	respMsg = []byte{}

	defer func() {
		if r := recover(); r != nil {
			err = closedConnErr
		}
	}()

	msgId := strconv.FormatInt(time.Now().UnixNano(), 16)
	msg := &Message{
		MessageType: messageType,
		MessageId:   msgId,
		Payload:     payload,
	}

	respCh := make(chan *MessageResp, 1)

	timeout := Config().WebSocketConnections.Timeouts.Request
	select {
	case <-time.After(timeout):
		err = errors.New(fmt.Sprintf("request timed out for %s", timeout))
		return
	case c.wch <- &MessageReq{
		msg:    msg,
		respCh: respCh,
	}:
	}

	if messageType == TypeRequestMessage {
		defer func() {
			c.msgChans.delete(msgId)
		}()

		timeout = Config().WebSocketConnections.Timeouts.Response
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

func (c *WebSocketConnection) wListen() {
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
				if req.msg.MessageType == TypeRequestMessage {
					c.msgChans.put(req.msg.MessageId, req.respCh)
				} else {
					req.respCh <- &MessageResp{}
				}
			}
		} else {
			<-c.closewch
			c.sendWsMessage(&Message{MessageType: TypeDisconnectMessage})
			return
		}
	}
}

func (c *WebSocketConnection) sendWsMessage(message *Message) error {
	err := c.ws.SetWriteDeadline(time.Now().Add(Config().WebSocketConnections.Timeouts.Write))
	if err != nil {
		return &WebsocketError{message: "error setting write deadline, " + err.Error()}
	}

	if message.MessageType == TypeDisconnectMessage {
		err = c.ws.WriteMessage(websocket.CloseMessage, message.Payload)
		err = c.ws.Close()
	} else {
		var p []byte
		p, err = message.Marshal()
		if err == nil {
			err = c.ws.WriteMessage(websocket.BinaryMessage, p)
			if err != nil {
				err = &WebsocketError{message: err.Error()}
			}
		}
	}
	return err
}

func (c *WebSocketConnection) readWsMessage() (*Message, error) {
	if err := c.ws.SetReadDeadline(time.Now().Add(Config().WebSocketConnections.Timeouts.Read)); err != nil {
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
			MessageType: TypeDisconnectMessage,
		}, nil
	}

	return Unmarshal(messageBody)
}

func (c *WebSocketConnection) rListen() {
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
			} else if message.MessageType == TypeDisconnectMessage {
				c.Close()
				return
			} else if message.MessageType == TypeResponseMessage {
				ch, found := c.msgChans.find(message.MessageId)
				if found {
					c.msgChans.delete(message.MessageId)
					ch <- &MessageResp{msg: message}
					c.h(c, message, nil)
				} else {
					c.h(c, message, unexpectedMessageErr)
				}
			} else {
				c.h(c, message, nil)
			}
		}
	}
}

func (c *WebSocketConnection) Close() {
	c.closeOnce.Do(func() {
		c.closed = true
		c.closedAt = time.Now()
		close(c.wch)
		close(c.rch)
		c.closewch <- true
		c.shard.unregister(c)
		Logger.Debug(fmt.Sprintf("connection: %s closed", c.Identifier()))
		c.h(c, &Message{MessageType: TypeDisconnectMessage}, nil)
	})
}

func (c *WebSocketConnection) Wait() {
	c.wg.Wait()
}

func (c *WebSocketConnection) Start() {
	c.wg.Add(2)
	go c.rListen()
	go c.wListen()
	Logger.Debug(fmt.Sprintf("connection: %s started", c.Identifier()))
	c.h(c, &Message{MessageType: TypeConnectMessage}, nil)
}
