package connections

import (
	"sync"
)

type MessageHandler func(Connection, Message, error)

type Middleware struct {
	name        string
	handlerFunc func(MessageHandler) MessageHandler
}

type MiddlewareStack struct {
	m           sync.Mutex
	middlewares []*Middleware
}

// Not thread safe!
// This is implimented without thread safty intentionally.
// We don't encouge using this method concurrently with other methods
func (ms *MiddlewareStack) Chain(h MessageHandler) MessageHandler {
	// Most inner handler should be empty.
	if h == nil {
		h = func(Connection, Message, error) {}
	}

	// Pop out handlers from middleware stack and wrap them togther such
	// that the bottom handler in stack will be the most outer handler which
	// gets executed first. The chain of execution ends at the empty handler.
	for i := len(ms.middlewares) - 1; i >= 0; i-- {
		h = ms.middlewares[i].handlerFunc(h)
	}
	return h
}

func (ms *MiddlewareStack) Use(m *Middleware) {
	ms.m.Lock()
	defer ms.m.Unlock()

	ms.middlewares = append(ms.middlewares, m)
}

func (ms *MiddlewareStack) InsertBefore(m, before *Middleware) {
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

func (ms *MiddlewareStack) findMiddleware(find *Middleware) int {
	idx := -1
	for i, m := range ms.middlewares {
		if m.name == find.name {
			idx = i
			break
		}
	}
	return idx
}

func (ms *MiddlewareStack) InsertAfter(m, after *Middleware) {
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

func (ms *MiddlewareStack) Remove(m *Middleware) {
	ms.m.Lock()
	defer ms.m.Unlock()

	idx := ms.findMiddleware(m)
	if idx != -1 {
		ms.middlewares = append(ms.middlewares[:idx], ms.middlewares[idx+1:]...)
	}
}

func NewMiddlewareStack() *MiddlewareStack {
	return &MiddlewareStack{middlewares: make([]*Middleware, 0)}
}

func NewMiddleware(name string, h func(MessageHandler) MessageHandler) *Middleware {
	return &Middleware{
		name:        name,
		handlerFunc: h,
	}
}
