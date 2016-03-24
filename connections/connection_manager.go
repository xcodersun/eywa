package connections

import (
	"errors"
	"github.com/vivowares/eywa/Godeps/_workspace/src/github.com/gorilla/websocket"
	. "github.com/vivowares/eywa/configs"
	"strconv"
	"sync"
	"time"
)

var closedCMErr = errors.New("connection manager is closed")

var HttpCloseChan = make(chan struct{})

type ConnectionManager struct {
	closed bool
	conns  map[string]Connection
	sync.Mutex
}

func (cm *ConnectionManager) NewWebsocketConnection(id, reqId string, ws wsConn, h MessageHandler, meta map[string]interface{}) (*WebsocketConnection, error) {

	conn := &WebsocketConnection{
		cm:           cm,
		ws:           ws,
		identifier:   id,
		createdAt:    time.Now(),
		lastPingedAt: time.Now(),
		h:            h,
		metadata:     meta,

		wch: make(chan *MessageReq, Config().Connections.Websocket.RequestQueueSize),
		msgChans: &syncRespChanMap{
			m: make(map[string]chan *MessageResp),
		},
		closewch: make(chan bool, 1),
		rch:      make(chan struct{}),
	}

	ws.SetPingHandler(func(payload string) error {
		conn.lastPingedAt = time.Now()
		//extend the read deadline after each ping
		err := ws.SetReadDeadline(time.Now().Add(Config().Connections.Websocket.Timeouts.Read.Duration))
		if err != nil {
			return err
		}

		return ws.WriteControl(
			websocket.PongMessage,
			[]byte(strconv.FormatInt(time.Now().UnixNano()/int64(time.Millisecond), 10)),
			time.Now().Add(Config().Connections.Websocket.Timeouts.Write.Duration))
	})

	cm.Lock()

	if cm.closed {
		ws.WriteControl(websocket.CloseMessage, []byte{}, time.Now().Add(Config().Connections.Websocket.Timeouts.Write.Duration))
		ws.Close()
		cm.Unlock()
		return nil, closedCMErr
	}

	_conn, found := cm.conns[conn.Identifier()]

	cm.conns[conn.Identifier()] = conn
	cm.Unlock()

	if found {
		go _conn.close(false)
	}

	conn.start()

	return conn, nil
}

func (cm *ConnectionManager) NewHttpConnection(id, reqId string, ch chan []byte, h MessageHandler, meta map[string]interface{}) (*HttpConnection, error) {
	if ch == nil {
		return &HttpConnection{
			requestId:  reqId,
			identifier: id,
			h:          h,
			metadata:   meta,
			createdAt:  time.Now(),
		}, nil
	}

	conn := &HttpConnection{
		requestId:  reqId,
		identifier: id,
		h:          h,
		ch:         ch,
		metadata:   meta,
		createdAt:  time.Now(),
		cm:         cm,
	}

	cm.Lock()
	if cm.closed {
		cm.Unlock()
		close(ch)
		return nil, closedCMErr
	}

	_conn, found := cm.conns[conn.Identifier()]

	cm.conns[conn.Identifier()] = conn
	cm.Unlock()

	if found {
		go _conn.close(false)
	}

	conn.start()
	return conn, nil
}

func (cm *ConnectionManager) FindConnection(id string) (Connection, bool) {
	cm.Lock()
	defer cm.Unlock()

	conn, found := cm.conns[id]
	return conn, found
}

func (cm *ConnectionManager) Count() int {
	cm.Lock()
	defer cm.Unlock()

	return len(cm.conns)
}

func (cm *ConnectionManager) close() error {
	cm.Lock()

	if cm.closed {
		cm.Unlock()
		return nil
	}

	cm.closed = true

	var wg sync.WaitGroup
	conns := make([]Connection, len(cm.conns))
	i := 0
	for _, conn := range cm.conns {
		conns[i] = conn
		i += 1
	}
	wg.Add(len(conns))

	cm.Unlock()

	for _, conn := range conns {
		go func(c Connection) {
			c.close(true)
			c.wait()
			wg.Done()
		}(conn)
	}

	wg.Wait()

	return nil
}

func (cm *ConnectionManager) unregister(c Connection) {
	cm.Lock()
	defer cm.Unlock()

	delete(cm.conns, c.Identifier())
}

func (cm *ConnectionManager) Closed() bool {
	return cm.closed
}
