package connections

import (
	"errors"
	"fmt"
	"github.com/vivowares/eywa/Godeps/_workspace/src/github.com/gorilla/websocket"
	. "github.com/vivowares/eywa/configs"
	. "github.com/vivowares/eywa/loggers"
	"io"
	"net"
	"strconv"
	"sync"
	"time"
)

var wsClosedConnErr = errors.New("websocket connection is closed")
var wsUnexpectedMessageErr = errors.New("unexpected response message received from websocket, probably due to response timeout?")

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

type WebsocketConnection struct {
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

func (c *WebsocketConnection) Identifier() string { return c.identifier }

func (c *WebsocketConnection) CreatedAt() time.Time { return c.createdAt }

func (c *WebsocketConnection) ClosedAt() time.Time { return c.closedAt }

func (c *WebsocketConnection) LastPingedAt() time.Time { return c.lastPingedAt }

func (c *WebsocketConnection) Closed() bool { return c.closed }

func (c *WebsocketConnection) MessageHandler() MessageHandler { return c.h }

func (c *WebsocketConnection) Metadata() map[string]interface{} { return c.metadata }

func (c *WebsocketConnection) Send(msg []byte) error {
	return c.sendAsyncMessage(TypeSendMessage, msg)
}

func (c *WebsocketConnection) Response(msg []byte) error {
	return c.sendAsyncMessage(TypeResponseMessage, msg)
}

func (c *WebsocketConnection) Request(msg []byte, timeout time.Duration) ([]byte, error) {
	return c.sendSyncMessage(TypeRequestMessage, msg, timeout)
}

func (c *WebsocketConnection) sendAsyncMessage(messageType int, payload []byte) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = wsClosedConnErr
		}
	}()

	msgId := strconv.FormatInt(time.Now().UnixNano(), 16)
	msg := &Message{
		MessageType: messageType,
		MessageId:   msgId,
		Payload:     payload,
	}

	respCh := make(chan *MessageResp, 1)

	timeout := Config().Connections.Websocket.Timeouts.Request.Duration
	select {
	case <-time.After(timeout):
		err = errors.New(fmt.Sprintf("request timed out for %s", timeout))
		return
	case c.wch <- &MessageReq{
		msg:    msg,
		respCh: respCh,
	}:
	}

	timeout = Config().Connections.Websocket.Timeouts.Request.Duration
	select {
	case <-time.After(timeout):
		err = errors.New(fmt.Sprintf("request timed out for %s", timeout))
		return
	case resp := <-respCh:
		err = resp.err
		return
	}
}

func (c *WebsocketConnection) sendSyncMessage(messageType int, payload []byte, timeout time.Duration) (respMsg []byte, err error) {
	respMsg = []byte{}

	defer func() {
		if r := recover(); r != nil {
			err = wsClosedConnErr
		}
	}()

	msgId := strconv.FormatInt(time.Now().UnixNano(), 16)
	msg := &Message{
		MessageType: messageType,
		MessageId:   msgId,
		Payload:     payload,
	}

	respCh := make(chan *MessageResp, 1)

	reqTimeout := Config().Connections.Websocket.Timeouts.Request.Duration
	select {
	case <-time.After(reqTimeout):
		err = errors.New(fmt.Sprintf("request timed out for %s", reqTimeout))
		return
	case c.wch <- &MessageReq{
		msg:    msg,
		respCh: respCh,
	}:
		defer func() {
			c.msgChans.delete(msgId)
		}()
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

func (c *WebsocketConnection) wListen() {
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

func (c *WebsocketConnection) sendWsMessage(message *Message) error {
	err := c.ws.SetWriteDeadline(time.Now().Add(Config().Connections.Websocket.Timeouts.Write.Duration))
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

func (c *WebsocketConnection) readWsMessage() (*Message, error) {
	if err := c.ws.SetReadDeadline(time.Now().Add(Config().Connections.Websocket.Timeouts.Read.Duration)); err != nil {
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
	// c.shard.updateRegistry(c)

	if messageType == websocket.CloseMessage {
		return &Message{
			MessageType: TypeDisconnectMessage,
		}, nil
	}

	return Unmarshal(messageBody)
}

func (c *WebsocketConnection) rListen() {
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
					c.h(c, message, wsUnexpectedMessageErr)
				}
			} else {
				c.h(c, message, nil)
			}
		}
	}
}

func (c *WebsocketConnection) Close() error {
	c.closeOnce.Do(func() {
		c.closed = true
		c.closedAt = time.Now()
		close(c.wch)
		close(c.rch)
		c.closewch <- true
		c.shard.unregister(c)
		Logger.Debug(fmt.Sprintf("websocket connection: %s closed", c.Identifier()))
		c.h(c, &Message{MessageType: TypeDisconnectMessage}, nil)
	})
	return nil
}

func (c *WebsocketConnection) Wait() {
	c.wg.Wait()
}

func (c *WebsocketConnection) Start() {
	c.wg.Add(2)
	go c.rListen()
	go c.wListen()
	Logger.Debug(fmt.Sprintf("websocket connection: %s started", c.Identifier()))
	c.h(c, &Message{MessageType: TypeConnectMessage}, nil)
}

func (c *WebsocketConnection) ConnectionType() string {
	return "websocket"
}
