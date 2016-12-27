package connections

import (
	"errors"
	"github.com/google/btree"
	"github.com/gorilla/websocket"
	. "github.com/eywa/configs"
	"github.com/eywa/pubsub"
	"strconv"
	"strings"
	"sync"
	"time"
)

var closedCMErr = errors.New("connection manager is closed")
var degree = 32

type ConnectionManager struct {
	id     string
	closed bool
	conns  *btree.BTree
	sync.Mutex
}

func (cm *ConnectionManager) Id() string { return cm.id }

func (cm *ConnectionManager) NewWebsocketConnection(id string, ws wsConn, h MessageHandler, meta map[string]string) (*WebsocketConnection, error) {
	p := pubsub.NewBasicPublisher(
		strings.Replace(cm.id, "/", "-", -1) + "/" +
			strings.Replace(id, "/", "-", -1))

	conn := &WebsocketConnection{
		cm:             cm,
		ws:             ws,
		identifier:     id,
		createdAt:      time.Now(),
		lastPingedAt:   time.Now(),
		h:              h,
		metadata:       meta,
		BasicPublisher: p,

		wch: make(chan *websocketMessageReq, Config().Connections.Websocket.RequestQueueSize),
		msgChans: &syncRespChanMap{
			m: make(map[string]chan *websocketMessageResp),
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
		cm.Unlock()
		ws.WriteControl(websocket.CloseMessage, []byte{}, time.Now().Add(Config().Connections.Websocket.Timeouts.Write.Duration))
		ws.Close()
		return nil, closedCMErr
	}

	_conn := cm.conns.ReplaceOrInsert(conn)

	cm.Unlock()

	if _conn != nil  {
		go _conn.(Connection).close(false)
	}

	conn.start()

	return conn, nil
}

func (cm *ConnectionManager) NewHttpConnection(id string, httpConn *httpConn, h MessageHandler, meta map[string]string) (*HttpConnection, error) {
	p := pubsub.NewBasicPublisher(
		strings.Replace(cm.id, "/", "-", -1) + "/" +
			strings.Replace(id, "/", "-", -1))

	conn := &HttpConnection{
		identifier:     id,
		h:              h,
		httpConn:       httpConn,
		metadata:       meta,
		createdAt:      time.Now(),
		cm:             cm,
		BasicPublisher: p,
	}

	conn.start()

	if httpConn._type == HttpPush {
		conn.close(false)
		return conn, nil
	}

	cm.Lock()
	if cm.closed {
		cm.Unlock()
		conn.close(false)
		return nil, closedCMErr
	}

	_conn := cm.conns.ReplaceOrInsert(conn)
	cm.Unlock()

	if _conn != nil {
		go _conn.(Connection).close(false)
	}

	return conn, nil
}

func (cm *ConnectionManager) FindConnection(id string) (Connection, bool) {
	cm.Lock()
	defer cm.Unlock()

	_conn := cm.conns.Get(&Lesser{id: id})
	if _conn == nil {
		return nil, false
	}

	return _conn.(Connection), true
}

func (cm *ConnectionManager) Count() int {
	cm.Lock()
	defer cm.Unlock()

	return cm.conns.Len()
}

func (cm *ConnectionManager) close() error {
	cm.Lock()

	if cm.closed {
		cm.Unlock()
		return nil
	}

	cm.closed = true

	var wg sync.WaitGroup
	conns := make([]Connection, cm.conns.Len())
	i := 0
	cm.conns.Ascend(func(it btree.Item) bool {
		conns[i] = it.(Connection)
		i += 1
		return true
	})
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

	cm.conns.Delete(&Lesser{id: c.Identifier()})
}

func (cm *ConnectionManager) Closed() bool {
	return cm.closed
}

func (cm *ConnectionManager) Scan(lastId string, size int) []Connection {
	cm.Lock()
	defer cm.Unlock()

	conns := make([]Connection, 0)
	if len(lastId) == 0 {
		i := 0
		cm.conns.Ascend(func(it btree.Item) bool {
			i += 1
			if i > size {
				return false
			}

			conns = append(conns, it.(Connection))
			return true
		})
	} else {
		i := 0
		cm.conns.AscendGreaterOrEqual(&Lesser{id: lastId}, func(it btree.Item) bool {
			conn := it.(Connection)
			if conn.Identifier() == lastId {
				return true
			} else {
				i += 1
				if i > size {
					return false
				}
				conns = append(conns, conn)
				return true
			}
		})
	}

	return conns
}
