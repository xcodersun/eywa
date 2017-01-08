package connections

import (
	"errors"
	"fmt"
	"github.com/google/btree"
	"github.com/gorilla/websocket"
	. "github.com/eywa/configs"
	"io/ioutil"
	"net/http"
	"sync"
)

var cmLock sync.RWMutex
var closed bool
var connManagers = make(map[string]*ConnectionManager)

var serverClosedErr = errors.New("server closed")

var WsUp *websocket.Upgrader
var HttpUp *HttpUpgrader

type HttpUpgrader struct{}

func (u *HttpUpgrader) Upgrade(w http.ResponseWriter, r *http.Request, _type HttpConnectionType) (*httpConn, error) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	if _type == HttpPush {
		return &httpConn{
			_type: HttpPush,
			body:  body,
		}, nil
	} else if _type == HttpPoll {
		return &httpConn{
			_type: HttpPoll,
			ch:    make(chan []byte, 1),
			body:  body,
		}, nil
	} else {
		return nil, errors.New(fmt.Sprintf("unsupported http connection type %d", _type))
	}
}

func InitWsUpgraders() {
	WsUp = &websocket.Upgrader{
		ReadBufferSize:  Config().Connections.Websocket.BufferSizes.Read,
		WriteBufferSize: Config().Connections.Websocket.BufferSizes.Write,
	}

	HttpUp = &HttpUpgrader{}
}

func InitializeCMs(ids []string) error {
	for _, id := range ids {
		_, err := NewConnectionManager(id)
		if err != nil {
			return err
		}
	}
	return nil
}

func CloseCMs() {
	ids := make([]string, 0)
	cmLock.Lock()
	closed = true
	for id, _ := range connManagers {
		ids = append(ids, id)
	}
	cmLock.Unlock()

	var wg sync.WaitGroup
	wg.Add(len(ids))
	for _, id := range ids {
		go func(_id string) {
			CloseConnectionManager(_id)
			wg.Done()
		}(id)
	}

	wg.Wait()
}

func NewConnectionManager(id string) (*ConnectionManager, error) {
	cm := &ConnectionManager{id: id, conns: btree.New(degree)}

	cmLock.Lock()
	defer cmLock.Unlock()

	if closed {
		return nil, serverClosedErr
	}

	if _, found := connManagers[id]; found {
		return nil, errors.New(fmt.Sprintf("connection manager %s already initialized.", id))
	}

	connManagers[id] = cm
	return cm, nil
}

func CloseConnectionManager(id string) error {
	cmLock.Lock()
	cm, found := connManagers[id]
	if !found {
		cmLock.Unlock()
		return errors.New(fmt.Sprintf("connection manager %s is not found", id))
	}
	delete(connManagers, id)
	cmLock.Unlock()

	return cm.close()
}

func Counts() (map[string]int, int) {
	cms := make(map[string]*ConnectionManager)
	cmLock.RLock()
	for id, cm := range connManagers {
		cms[id] = cm
	}
	cmLock.RUnlock()

	var total int = 0
	counts := make(map[string]int)
	for id, cm := range cms {
		counts[id] = cm.Count()
		total += cm.Count()
	}

	return counts, total
}

func FindConnectionManager(id string) (*ConnectionManager, bool) {
	cmLock.RLock()
	defer cmLock.RUnlock()
	cm, found := connManagers[id]
	return cm, found
}
