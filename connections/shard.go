package connections

import (
	"errors"
	"sync"
)

type shard struct {
	wscm     *WebSocketConnectionManager
	wsconns  map[string]*WebSocketConnection
	closed bool
	sync.Mutex
}

func (sh *shard) Close() {
	sh.Lock()

	sh.closed = true

	var wg sync.WaitGroup
	wsconns := make([]*WebSocketConnection, len(sh.wsconns))
	i := 0
	for _, conn := range sh.wsconns {
		wsconns[i] = conn
		i += 1
	}
	wg.Add(len(wsconns))

	sh.Unlock()

	for _, conn := range wsconns {
		go func(c *WebSocketConnection) {
			c.Close()
			c.Wait()
			wg.Done()
		}(conn)
	}

	wg.Wait()
}

func (sh *shard) register(c *WebSocketConnection) error {
	sh.Lock()
	defer sh.Unlock()

	if sh.closed {
		return errors.New("shard is closed")
	}

	if err := sh.wscm.Registry.Register(c); err != nil {
		return err
	}

	sh.wsconns[c.identifier] = c
	return nil
}

func (sh *shard) updateRegistry(c *WebSocketConnection) error {
	return sh.wscm.Registry.UpdateRegistry(c)
}

func (sh *shard) unregister(c *WebSocketConnection) error {
	sh.Lock()
	defer sh.Unlock()

	delete(sh.wsconns, c.identifier)
	return sh.wscm.Registry.Unregister(c)
}

func (sh *shard) findConnection(id string) (*WebSocketConnection, bool) {
	sh.Lock()
	defer sh.Unlock()

	conn, found := sh.wsconns[id]
	return conn, found
}

func (sh *shard) Count() int {
	sh.Lock()
	defer sh.Unlock()

	return len(sh.wsconns)
}
