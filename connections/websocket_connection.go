package connections

import (
	"errors"
	"fmt"
	"github.com/google/btree"
	"github.com/gorilla/websocket"
	. "github.com/eywa/configs"
	"github.com/eywa/pubsub"
	"io"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"
)

var wsConnClosedErr = errors.New("websocket connection is closed")
var wsUnexpectedMessageErr = errors.New("unexpected response message received from websocket connection, probably due to response timeout?")

type websocketError struct {
	message string
}

func (e *websocketError) Error() string {
	return fmt.Sprintf("WebsocketError: %s", e.message)
}

type syncRespChanMap struct {
	sync.Mutex
	m map[string]chan *websocketMessageResp
}

func (sm *syncRespChanMap) put(msgId string, ch chan *websocketMessageResp) {
	sm.Lock()
	defer sm.Unlock()

	sm.m[msgId] = ch
}

func (sm *syncRespChanMap) find(msgId string) (chan *websocketMessageResp, bool) {
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
	// Reference to the connection manager
	cm           *ConnectionManager
	// Gorilla websocket
	ws           wsConn
	// Websocket connection creation time. This can be used to compute
	// total connection alive time for stats
	createdAt    time.Time
	// Last ping time. This will be used to check if client is alive
	// because in most cases, client will close it's socket unexpectedly.
	// TODO(alexsun): currently, it is not used yet.
	lastPingedAt time.Time
	// Websocket connection closing time. This can be used to compute
	// total connection alive time for stats.
	closedAt     time.Time
	// The key for connecton manager to find the websocket connection.
	// TODO(alexsun): currently, device id is used the key which could
	// be duplicated across different users. The key hould include user's
	// identification as well.
	identifier   string
	// The head message handler.
	h            MessageHandler
	metadata     map[string]string
	*pubsub.BasicPublisher

	// Write channel for wListen thread. By closing this channel does terminate
	// wListen thread.
	wch      chan *websocketMessageReq // size=?
	// Read channel for rListen thread. By closing this channel does terminate
	// rListen thread.
	rch      chan struct{}             // size=0
	closewch chan bool                 // size=1

	// Sync group for threads of rListen and wListen. It's incremented before
	// rListen and wListen starts and decremented when they are closed. The
	// wait() function will block caller until both rListen and wListen is
	// terminated. It is used by ConnectionManager's close() function which
	// need to make sure that each connection has closed its rListen and wListen
	// threads.
	rwStart   sync.WaitGroup
	// Make sure that the websocket connection is closed once and only once.
	closeOnce sync.Once
	closed    bool

	// There is a chance for this msgChan to grow,
	// in extreme race condition. no plan to fix it.
	// simple solution is to limit the size of it,
	// close the connection when it blows up.
	msgChans *syncRespChanMap
}

func (c *WebsocketConnection) Identifier() string { return c.identifier }

func (c *WebsocketConnection) CreatedAt() time.Time { return c.createdAt }

func (c *WebsocketConnection) ClosedAt() time.Time { return c.closedAt }

func (c *WebsocketConnection) LastPingedAt() time.Time { return c.lastPingedAt }

func (c *WebsocketConnection) Closed() bool { return c.closed }

func (c *WebsocketConnection) Metadata() map[string]string { return c.metadata }

func (c *WebsocketConnection) ConnectionManager() *ConnectionManager { return c.cm }

func (c *WebsocketConnection) Less(than btree.Item) bool {
	conn := than.(Connection)
	return strings.Compare(c.identifier, conn.Identifier()) < 0
}

func (c *WebsocketConnection) Send(msg []byte) error {
	return c.sendAsyncMessage(TypeSendMessage, msg)
}

func (c *WebsocketConnection) Request(msg []byte, timeout time.Duration) ([]byte, error) {
	return c.sendSyncMessage(TypeRequestMessage, msg, timeout)
}

func (c *WebsocketConnection) sendAsyncMessage(messageType MessageType, payload []byte) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = wsConnClosedErr
		}
	}()

	msg := &websocketMessage{
		_type:   messageType,
		payload: payload,
	}

	respCh := make(chan *websocketMessageResp, 1)

	timeout := Config().Connections.Websocket.Timeouts.Request.Duration
	select {
	case <-time.After(timeout):
		err = errors.New(fmt.Sprintf("websocket connection request timed out for %s", timeout))
		return
	case c.wch <- &websocketMessageReq{
		msg:    msg,
		respCh: respCh,
	}:
	}

	timeout = Config().Connections.Websocket.Timeouts.Request.Duration
	select {
	case <-time.After(timeout):
		err = errors.New(fmt.Sprintf("websocket connection request timed out for %s", timeout))
		return
	case resp := <-respCh:
		err = resp.err
		return
	}
}

func (c *WebsocketConnection) sendSyncMessage(messageType MessageType, payload []byte, timeout time.Duration) (respMsg []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = wsConnClosedErr
		}
	}()

	respMsg = []byte{}
	msg := &websocketMessage{
		_type:   messageType,
		id:      strconv.FormatInt(time.Now().UnixNano(), 16),
		payload: payload,
	}

	respCh := make(chan *websocketMessageResp, 1)

	reqTimeout := Config().Connections.Websocket.Timeouts.Request.Duration
	select {
	case <-time.After(reqTimeout):
		err = errors.New(fmt.Sprintf("websocket connection request timed out for %s", reqTimeout))
		return
	case c.wch <- &websocketMessageReq{
		msg:    msg,
		respCh: respCh,
	}:
		defer func() {
			c.msgChans.delete(msg.id)
		}()
	}

	select {
	case <-time.After(timeout):
		err = errors.New(fmt.Sprintf("websocket connection response timed out for %s", timeout))
		return
	case resp := <-respCh:
		if resp.msg != nil {
			respMsg = resp.msg.payload
		}
		err = resp.err
		return
	}

}

func (c *WebsocketConnection) wListen() {
	defer c.rwStart.Done()
	for {
		req, more := <-c.wch
		if more {
			err := c.sendWsMessage(req.msg)

			go c.h(c, req.msg, err)

			if err != nil {
				req.respCh <- &websocketMessageResp{
					msg: nil,
					err: err,
				}
				if _, ok := err.(*websocketError); ok {
					c.close(true)
				}
			} else {
				if req.msg._type == TypeRequestMessage {
					c.msgChans.put(req.msg.id, req.respCh)
				} else {
					req.respCh <- &websocketMessageResp{}
				}
			}

		} else {
			<-c.closewch
			c.sendWsMessage(&websocketMessage{_type: TypeDisconnectMessage})
			return
		}
	}
}

func (c *WebsocketConnection) sendWsMessage(message *websocketMessage) (err error) {
	var p []byte
	p, err = message.Marshal()
	if err != nil {
		return
	}

	err = c.ws.SetWriteDeadline(time.Now().Add(Config().Connections.Websocket.Timeouts.Write.Duration))
	if err != nil {
		return &websocketError{message: "error setting write deadline for websocket connection, " + err.Error()}
	}

	if message._type == TypeDisconnectMessage {
		err = c.ws.WriteMessage(websocket.CloseMessage, message.payload)
		err = c.ws.Close()
	} else {
		err = c.ws.WriteMessage(websocket.BinaryMessage, p)
		if err != nil {
			err = &websocketError{message: err.Error()}
		}
	}
	return
}

func (c *WebsocketConnection) readWsMessage() (*websocketMessage, error) {
	if err := c.ws.SetReadDeadline(time.Now().Add(Config().Connections.Websocket.Timeouts.Read.Duration)); err != nil {
		return nil, &websocketError{
			message: fmt.Sprintf("error setting read deadline for websocket connection, %s", err.Error()),
		}
	}

	messageType, messageBody, err := c.ws.ReadMessage()
	if err != nil {
		return nil, &websocketError{
			message: fmt.Sprintf("error reading message from websocket connection, %s", err.Error()),
		}
	}

	c.lastPingedAt = time.Now()

	if messageType == websocket.CloseMessage {
		return &websocketMessage{
			_type: TypeDisconnectMessage,
		}, nil
	}

	m := &websocketMessage{raw: messageBody}
	err = m.Unmarshal()
	return m, err
}

func (c *WebsocketConnection) rListen() {
	defer c.rwStart.Done()
	for {
		select {
		case <-c.rch:
			return
		default:
			message, err := c.readWsMessage()
			if err != nil {
				if _, ok := err.(*websocketError); ok {
					go c.h(c, nil, err)
					c.close(true)
					return
				}
				go c.h(c, message, err)
			} else if message._type == TypeDisconnectMessage {
				go c.h(c, message, nil)
				c.close(true)
				return
			} else if message._type == TypeResponseMessage {
				ch, found := c.msgChans.find(message.id)
				if found {
					c.msgChans.delete(message.id)
					ch <- &websocketMessageResp{msg: message}
					go c.h(c, message, nil)
				} else {
					go c.h(c, message, wsUnexpectedMessageErr)
				}
			} else {
				go c.h(c, message, nil)
			}
		}
	}
}

func (c *WebsocketConnection) unregister() {
	// To avoid race condition where a new connection has registered
	// under the same id and current connection become orphan, in which
	// case the orphan connection has different creatd time with the
	// registered connection
	conn, found := c.cm.FindConnection(c.identifier)
	if found && conn.CreatedAt() == c.createdAt {
		c.cm.unregister(c)
	}
}

func (c *WebsocketConnection) close(unregister bool) error {
	c.closeOnce.Do(func() {
		c.closed = true
		c.closedAt = time.Now()
		close(c.wch)
		close(c.rch)
		c.closewch <- true
		if unregister {
			c.unregister()
		}
		go c.h(c, &websocketMessage{_type: TypeDisconnectMessage}, nil)
		go func() {
			time.Sleep(3 * time.Second) // for user experience
			c.BasicPublisher.Unpublish()
		}()
	})
	return nil
}

func (c *WebsocketConnection) wait() {
	c.rwStart.Wait()
}

func (c *WebsocketConnection) start() {
	// Start listening to read channel
	c.rwStart.Add(1)
	go c.rListen()
	// Start listening to write channel
	c.rwStart.Add(1)
	go c.wListen()
	// Start the message handler chain
	go c.h(c, &websocketMessage{_type: TypeConnectMessage}, nil)
}

func (c *WebsocketConnection) ConnectionType() string {
	return "websocket"
}
