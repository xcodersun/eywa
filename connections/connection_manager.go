package connections

import (
	"errors"
	"fmt"
	"github.com/vivowares/eywa/Godeps/_workspace/src/github.com/gorilla/websocket"
	"github.com/vivowares/eywa/Godeps/_workspace/src/github.com/spaolacci/murmur3"
	. "github.com/vivowares/eywa/configs"
	. "github.com/vivowares/eywa/loggers"
	. "github.com/vivowares/eywa/utils"
	"strconv"
	"sync"
	"time"
)

var closedCMErr = errors.New("connection manager is closed")

var HttpCloseChan = make(chan struct{})

type ConnectionManager struct {
	closed   *AtomBool
	shards   []*shard
	Registry Registry
}

func (cm *ConnectionManager) NewWebsocketConnection(id string, ws wsConn, h MessageHandler, meta map[string]interface{}) (*WebsocketConnection, error) {
	if cm.closed.Get() {
		ws.Close()
		return nil, closedCMErr
	}

	hasher := murmur3.New32()
	hasher.Write([]byte(id))
	shard := cm.shards[hasher.Sum32()%uint32(len(cm.shards))]

	t := time.Now()
	conn := &WebsocketConnection{
		shard:        shard,
		ws:           ws,
		identifier:   id,
		createdAt:    t,
		lastPingedAt: t,
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
		// conn.shard.updateRegistry(conn)
		Logger.Debug(fmt.Sprintf("websocket connection: %s pinged", id))

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

	if err := shard.register(conn); err != nil {
		ws.WriteControl(websocket.CloseMessage, []byte{}, time.Now().Add(Config().Connections.Websocket.Timeouts.Write.Duration))
		ws.Close()
		return nil, err
	}

	conn.Start()

	return conn, nil
}

func (cm *ConnectionManager) NewHttpConnection(id string, ch chan []byte, h MessageHandler, meta map[string]interface{}) (*HttpConnection, error) {
	if ch == nil {
		return &HttpConnection{
			identifier: id,
			h:          h,
			metadata:   meta,
			createdAt:  time.Now(),
		}, nil
	}

	if cm.closed.Get() {
		close(ch)
		return nil, closedCMErr
	}

	hasher := murmur3.New32()
	hasher.Write([]byte(id))
	shard := cm.shards[hasher.Sum32()%uint32(len(cm.shards))]

	t := time.Now()
	conn := &HttpConnection{
		identifier: id,
		h:          h,
		ch:         ch,
		metadata:   meta,
		createdAt:  t,
		shard:      shard,
	}

	if err := shard.register(conn); err != nil {
		close(ch)
		return nil, err
	}

	conn.Start()

	return conn, nil
}

func (cm *ConnectionManager) FindConnection(id string) (Connection, bool) {
	hasher := murmur3.New32()
	hasher.Write([]byte(id))
	shard := cm.shards[hasher.Sum32()%uint32(len(cm.shards))]
	return shard.findConnection(id)
}

func (cm *ConnectionManager) Count() int {
	sum := 0
	for _, sh := range cm.shards {
		sum += sh.count()
	}
	return sum
}

func (cm *ConnectionManager) close() error {
	cm.closed.Set(true)

	var wg sync.WaitGroup
	wg.Add(len(cm.shards))
	for _, sh := range cm.shards {
		go func(s *shard) {
			s.Close()
			wg.Done()
		}(sh)
	}
	wg.Wait()
	return cm.Registry.Close()
}
