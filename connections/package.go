package connections

import (
	"errors"
	"fmt"
	"sync"
)

var cmLock sync.RWMutex
var closed bool
var connManagers = make(map[string]*ConnectionManager)

var serverClosedErr = errors.New("server closed")

func InitializeCMs(names []string) error {
	for _, name := range names {
		_, err := NewConnectionManager(name)
		if err != nil {
			return err
		}
	}
	return nil
}

func CloseCMs() {
	names := make([]string, 0)
	cmLock.Lock()
	closed = true
	for name, _ := range connManagers {
		names = append(names, name)
	}
	cmLock.Unlock()

	var wg sync.WaitGroup
	wg.Add(len(names))
	for _, name := range names {
		go func(_name string) {
			CloseConnectionManager(_name)
			wg.Done()
		}(name)
	}

	wg.Wait()
}

func NewConnectionManager(name string) (*ConnectionManager, error) {
	cm := &ConnectionManager{conns: make(map[string]Connection)}

	cmLock.Lock()
	defer cmLock.Unlock()

	if closed {
		return nil, serverClosedErr
	}

	if _, found := connManagers[name]; found {
		return nil, errors.New(fmt.Sprintf("connection manager: %s already initialized.", name))
	}

	connManagers[name] = cm
	return cm, nil
}

func CloseConnectionManager(name string) error {
	cmLock.Lock()
	cm, found := connManagers[name]
	if !found {
		cmLock.Unlock()
		return errors.New(fmt.Sprintf("connection manager: %s is not found", name))
	}
	delete(connManagers, name)
	cmLock.Unlock()

	return cm.close()
}

func Counts() map[string]int {
	conns := make(map[string]*ConnectionManager)
	cmLock.RLock()
	for name, cm := range connManagers {
		conns[name] = cm
	}
	cmLock.RUnlock()

	counts := make(map[string]int)
	for name, cm := range conns {
		counts[name] = cm.Count()
	}

	return counts
}

func FindConnectionManager(name string) (*ConnectionManager, bool) {
	cmLock.RLock()
	defer cmLock.RUnlock()
	cm, found := connManagers[name]
	return cm, found
}
