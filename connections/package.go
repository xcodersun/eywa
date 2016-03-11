package connections

import (
	"errors"
	. "github.com/vivowares/eywa/configs"
	. "github.com/vivowares/eywa/utils"
	"time"
)

var defaultCM *ConnectionManager
var noDefaultCMErr = errors.New("default connection manager is not initialized")

func InitializeCM() error {
	cm, err := NewConnectionManager()
	if err == nil {
		defaultCM = cm
	}
	return err
}

func CloseCM() error {
	if defaultCM != nil {
		return defaultCM.Close()
	}

	return nil
}

func NewConnectionManager() (*ConnectionManager, error) {
	cm := &ConnectionManager{closed: &AtomBool{}}
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
			conns: make(map[string]Connection, Config().Connections.InitShardSize),
		}
	}

	return cm, nil
}

func Count() int {
	count := 0

	if defaultCM != nil {
		count = defaultCM.Count()
	}

	return count
}

func FindConnection(id string) (Connection, bool) {
	if defaultCM != nil {
		return defaultCM.FindConnection(id)
	}

	return nil, false
}

func NewHttpConnection(id string, ch chan []byte, h MessageHandler, meta map[string]interface{}) (*HttpConnection, error) {
	if ch != nil {
		return defaultCM.NewHttpConnection(id, ch, h, meta)
	} else {
		return &HttpConnection{
			identifier: id,
			h:          h,
			metadata:   meta,
			createdAt:  time.Now(),
		}, nil
	}
}

func NewWebsocketConnection(id string, ws wsConn, h MessageHandler, meta map[string]interface{}) (*WebsocketConnection, error) {
	if defaultCM != nil {
		return defaultCM.NewWebsocketConnection(id, ws, h, meta)
	}

	return nil, noDefaultCMErr
}
