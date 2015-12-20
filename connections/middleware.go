package connections

import (
	"sync"
)

type MessageHandler func(*Connection, *Message, error)

type Middleware struct {
	name       string
	middleware func(MessageHandler) MessageHandler
}

type Middlewares struct {
	m           sync.Mutex
	middlewares []*Middleware
}

// Not thread safe!
// This is implimented without thread safty intentionally.
// We don't encouge using this method concurrently with other methods
func (ms *Middlewares) Chain(h MessageHandler) MessageHandler {
	if h == nil {
		h = func(*Connection, *Message, error) {}
	}

	for i := len(ms.middlewares) - 1; i >= 0; i-- {
		h = ms.middlewares[i].middleware(h)
	}
	return h
}

func (ms *Middlewares) Use(m *Middleware) {
	ms.m.Lock()
	defer ms.m.Unlock()

	ms.middlewares = append(ms.middlewares, m)
}

func (ms *Middlewares) InsertBefore(m, before *Middleware) {
	ms.m.Lock()
	defer ms.m.Unlock()

	idx := ms.findMiddleware(before)
	if idx == -1 {
		ms.middlewares = append([]*Middleware{m}, ms.middlewares...)
	} else {
		ms.middlewares = append(ms.middlewares, m)
		copy(ms.middlewares[idx+1:], ms.middlewares[idx:])
		ms.middlewares[idx] = m
	}
}

func (ms *Middlewares) findMiddleware(find *Middleware) int {
	idx := -1
	for i, m := range ms.middlewares {
		if m.name == find.name {
			idx = i
			break
		}
	}
	return idx
}

func (ms *Middlewares) InsertAfter(m, after *Middleware) {
	ms.m.Lock()
	defer ms.m.Unlock()

	idx := ms.findMiddleware(after)
	if idx == -1 {
		ms.middlewares = append(ms.middlewares, m)
	} else {
		ms.middlewares = append([]*Middleware{m}, ms.middlewares...)
		copy(ms.middlewares[:idx+1], ms.middlewares[1:idx+2])
		ms.middlewares[idx+1] = m
	}
}

func (ms *Middlewares) Remove(m *Middleware) {
	ms.m.Lock()
	defer ms.m.Unlock()

	idx := ms.findMiddleware(m)
	if idx != -1 {
		ms.middlewares = append(ms.middlewares[:idx], ms.middlewares[idx+1:]...)
	}
}
