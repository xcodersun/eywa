package connections

import (
	"github.com/vivowares/octopus/Godeps/_workspace/src/github.com/gorilla/websocket"
	"github.com/vivowares/octopus/Godeps/_workspace/src/github.com/spaolacci/murmur3"
	. "github.com/vivowares/octopus/configs"
	"strconv"
	"sync"
	"time"
)

var CM *ConnectionManager

func InitializeCM() error {
	cm, err := NewConnectionManager()
	CM = cm
	return err
}

func NewConnectionManager() (*ConnectionManager, error) {
	cm := &ConnectionManager{}
	switch Config().Connections.Registry {
	case "memory":
		cm.Registry = &InMemoryRegistry{}
	default:
		cm.Registry = &InMemoryRegistry{}
	}
	if err := cm.Registry.Ping(); err != nil {
		return nil, err
	}

	cm.shards = make([]*shard, Config().Connections.NShards)
	for i := 0; i < Config().Connections.NShards; i++ {
		cm.shards[i] = &shard{
			cm:    cm,
			conns: make(map[string]*Connection, Config().Connections.InitShardSize),
		}
	}

	return cm, nil
}

type ConnectionManager struct {
	shards   []*shard
	Registry Registry
}

func (cm *ConnectionManager) Close() error {
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

func (cm *ConnectionManager) NewConnection(id string, ws wsConn, h MessageHandler, meta map[string]interface{}) (*Connection, error) {
	hasher := murmur3.New32()
	hasher.Write([]byte(id))
	shard := cm.shards[hasher.Sum32()%uint32(len(cm.shards))]

	t := time.Now()
	conn := &Connection{
		shard:        shard,
		ws:           ws,
		identifier:   id,
		createdAt:    t,
		lastPingedAt: t,
		h:            h,
		Metadata:     meta,

		wch: make(chan *MessageReq, Config().Connections.RequestQueueSize),
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
			time.Now().Add(Config().Connections.Timeouts.Write))
	})

	conn.Start()
	if err := shard.register(conn); err != nil {
		conn.Close()
		conn.Wait()
		return nil, err
	}

	return conn, nil
}

func (cm *ConnectionManager) FindConnection(id string) (*Connection, bool) {
	hasher := murmur3.New32()
	hasher.Write([]byte(id))
	shard := cm.shards[hasher.Sum32()%uint32(len(cm.shards))]
	return shard.findConnection(id)
}

func (cm *ConnectionManager) Count() int {
	sum := 0
	for _, sh := range cm.shards {
		sum += sh.Count()
	}
	return sum
}
