package connections

import (
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type ConnectionManager interface {
	Start() error
	Close() error
	Host() string
	NewConnection(string, *websocket.Conn, MessageHandler) (Connection, error)

	registerConnection(Connection) error
	unregisterConnection(Connection) error
	refreshConnectionRegistry(Connection, time.Time) error
}

type InMemoryConnectionManager struct {
	host        string
	connections map[string]Connection
	closed      bool

	wg sync.WaitGroup
	m  sync.RWMutex
}

func (cm *InMemoryConnectionManager) Host() string {
	return cm.host
}

func (cm *InMemoryConnectionManager) Start() error {
	cm.m.Lock()
	defer cm.m.Unlock()

	if len(cm.host) == 0 {
		return &ConnectionManagerStartingError{
			message: "empty host name",
		}
	}

	if !cm.closed {
		return &ConnectionManagerStartingError{
			message: "connection manager is already started",
		}
	}

	cm.connections = make(map[string]Connection)
	cm.closed = false
	return nil
}

func (cm *InMemoryConnectionManager) NewConnection(identifier string, ws *websocket.Conn, h MessageHandler) (Connection, error) {
	t := time.Now()
	conn := &connection{
		identifier:   identifier,
		createdAt:    t,
		lastPingedAt: t,
		closed:       false,
		closeChan:    make(chan bool, 1),
		msgChans:     make(map[string]chan *Message),
		ws:           ws,
		cm:           cm,
	}
	ws.SetPingHandler(func(payload string) error {
		conn.lastPingedAt = time.Now()
		return nil
	})

	if err := cm.registerConnection(conn); err != nil {
		conn.close()
		return nil, err
	}

	go conn.listen(h)
	return conn, nil
}

func (cm *InMemoryConnectionManager) registerConnection(conn Connection) error {
	cm.wg.Add(1)

	cm.m.Lock()
	defer cm.m.Unlock()

	if cm.closed {
		return &ConnectionRegisterError{
			message: "connection manager closed",
		}
	}

	cm.connections[conn.Identifier()] = conn

	return nil
}

func (cm *InMemoryConnectionManager) unregisterConnection(conn Connection) error {
	cm.m.Lock()

	if c, found := cm.connections[conn.Identifier()]; found {
		if conn.CreatedAt().Before(c.CreatedAt()) || conn.CreatedAt().Equal(c.CreatedAt()) {
			delete(cm.connections, conn.Identifier())
		}
	}

	cm.m.Unlock()

	cm.wg.Done()

	return nil
}

func (cm *InMemoryConnectionManager) refreshConnectionRegistry(conn Connection, t time.Time) error {
	return nil
}

func (cm *InMemoryConnectionManager) Close() error {
	cm.m.Lock()
	defer cm.wg.Wait()
	defer cm.m.Unlock()

	if cm.closed {
		return nil
	}

	cm.closed = true

	for _, c := range cm.connections {
		c.signalClose()
	}

	return nil
}
