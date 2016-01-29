package connections

import (
	"errors"
	"sync"
)

type shard struct {
	wscm     *WebSocketConnectionManager
	conns  map[string]*Connection
	closed bool
	sync.Mutex
}

func (sh *shard) Close() {
	sh.Lock()

	sh.closed = true

	var wg sync.WaitGroup
	conns := make([]*Connection, len(sh.conns))
	i := 0
	for _, conn := range sh.conns {
		conns[i] = conn
		i += 1
	}
	wg.Add(len(conns))

	sh.Unlock()

	for _, conn := range conns {
		go func(c *Connection) {
			c.Close()
			c.Wait()
			wg.Done()
		}(conn)
	}

	wg.Wait()
}

func (sh *shard) register(c *Connection) error {
	sh.Lock()
	defer sh.Unlock()

	if sh.closed {
		return errors.New("shard is closed")
	}

	if err := sh.wscm.Registry.Register(c); err != nil {
		return err
	}

	sh.conns[c.identifier] = c
	return nil
}

func (sh *shard) updateRegistry(c *Connection) error {
	return sh.wscm.Registry.UpdateRegistry(c)
}

func (sh *shard) unregister(c *Connection) error {
	sh.Lock()
	defer sh.Unlock()

	delete(sh.conns, c.identifier)
	return sh.wscm.Registry.Unregister(c)
}

func (sh *shard) findConnection(id string) (*Connection, bool) {
	sh.Lock()
	defer sh.Unlock()

	conn, found := sh.conns[id]
	return conn, found
}

func (sh *shard) Count() int {
	sh.Lock()
	defer sh.Unlock()

	return len(sh.conns)
}
