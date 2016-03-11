package connections

import (
	"sync"
)

type shard struct {
	cm    *ConnectionManager
	conns map[string]Connection
	sync.Mutex
}

func (sh *shard) Close() {
	sh.Lock()

	var wg sync.WaitGroup
	conns := make([]Connection, len(sh.conns))
	i := 0
	for _, conn := range sh.conns {
		conns[i] = conn
		i += 1
	}
	wg.Add(len(conns))

	sh.Unlock()

	for _, conn := range conns {
		go func(c Connection) {
			c.Close()
			c.Wait()
			wg.Done()
		}(conn)
	}

	wg.Wait()
}

func (sh *shard) register(c Connection) error {
	sh.Lock()
	defer sh.Unlock()

	if sh.cm.closed.Get() {
		return closedCMErr
	}

	if err := sh.cm.Registry.Register(c); err != nil {
		return err
	}

	sh.conns[c.Identifier()] = c
	return nil
}

func (sh *shard) unregister(c Connection) error {
	sh.Lock()
	defer sh.Unlock()

	delete(sh.conns, c.Identifier())
	return sh.cm.Registry.Unregister(c)
}

// func (sh *shard) updateRegistry(c Connection) error {
// 	return sh.cm.Registry.UpdateRegistry(c)
// }

func (sh *shard) findConnection(id string) (Connection, bool) {
	sh.Lock()
	defer sh.Unlock()

	conn, found := sh.conns[id]
	return conn, found
}

func (sh *shard) count() int {
	sh.Lock()
	defer sh.Unlock()

	return len(sh.conns)
}
