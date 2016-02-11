package connections

import (
	"github.com/vivowares/octopus/Godeps/_workspace/src/github.com/gorilla/websocket"
	"github.com/vivowares/octopus/Godeps/_workspace/src/github.com/spaolacci/murmur3"
	. "github.com/vivowares/octopus/configs"
	"strconv"
	"sync"
	"time"
)

var defaultWSCM *WebSocketConnectionManager

func InitializeWSCM() error {
	wscm, err := newWebSocketConnectionManager()
	defaultWSCM = wscm
	return err
}

func CloseWSCM() error {
	if defaultWSCM != nil {
		return defaultWSCM.close()
	}

	return NoWscmErr
}

func newWebSocketConnectionManager() (*WebSocketConnectionManager, error) {
	wscm := &WebSocketConnectionManager{}
	switch Config().WebSocketConnections.Registry {
	case "memory":
		wscm.Registry = &InMemoryRegistry{}
	default:
		wscm.Registry = &InMemoryRegistry{}
	}
	if err := wscm.Registry.Ping(); err != nil {
		return nil, err
	}

	wscm.shards = make([]*shard, Config().WebSocketConnections.NShards)
	for i := 0; i < Config().WebSocketConnections.NShards; i++ {
		wscm.shards[i] = &shard{
			wscm:    wscm,
			wsconns: make(map[string]*WebSocketConnection, Config().WebSocketConnections.InitShardSize),
		}
	}

	return wscm, nil
}

type WebSocketConnectionManager struct {
	shards   []*shard
	Registry Registry
}

func (wscm *WebSocketConnectionManager) close() error {
	var wg sync.WaitGroup
	wg.Add(len(wscm.shards))
	for _, sh := range wscm.shards {
		go func(s *shard) {
			s.Close()
			wg.Done()
		}(sh)
	}
	wg.Wait()
	return wscm.Registry.Close()
}

func (wscm *WebSocketConnectionManager) newConnection(id string, ws wsConn, h MessageHandler, meta map[string]interface{}) (*WebSocketConnection, error) {
	hasher := murmur3.New32()
	hasher.Write([]byte(id))
	shard := wscm.shards[hasher.Sum32()%uint32(len(wscm.shards))]

	t := time.Now()
	conn := &WebSocketConnection{
		shard:        shard,
		ws:           ws,
		identifier:   id,
		createdAt:    t,
		lastPingedAt: t,
		h:            h,
		metadata:     meta,

		wch: make(chan *MessageReq, Config().WebSocketConnections.RequestQueueSize),
		msgChans: &syncRespChanMap{
			m: make(map[string]chan *MessageResp),
		},
		closewch: make(chan bool, 1),
		rch:      make(chan struct{}),
	}

	ws.SetPingHandler(func(payload string) error {
		conn.lastPingedAt = time.Now()
		conn.shard.updateRegistry(conn)
		return ws.WriteControl(
			websocket.PongMessage,
			[]byte(strconv.FormatInt(time.Now().UnixNano()/int64(time.Millisecond), 10)),
			time.Now().Add(Config().WebSocketConnections.Timeouts.Write))
	})

	conn.Start()
	if err := shard.register(conn); err != nil {
		conn.Close()
		conn.Wait()
		return nil, err
	}

	return conn, nil
}

func (wscm *WebSocketConnectionManager) findConnection(id string) (*WebSocketConnection, bool) {
	hasher := murmur3.New32()
	hasher.Write([]byte(id))
	shard := wscm.shards[hasher.Sum32()%uint32(len(wscm.shards))]
	return shard.findConnection(id)
}

func (wscm *WebSocketConnectionManager) count() int {
	sum := 0
	for _, sh := range wscm.shards {
		sum += sh.Count()
	}
	return sum
}
