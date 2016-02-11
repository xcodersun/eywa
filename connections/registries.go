package connections

type Registry interface {
	Register(*WebSocketConnection) error
	Unregister(*WebSocketConnection) error
	UpdateRegistry(*WebSocketConnection) error
	Ping() error
	Close() error
}

type InMemoryRegistry struct{}

func (r *InMemoryRegistry) Register(c *WebSocketConnection) error       { return nil }
func (r *InMemoryRegistry) Unregister(c *WebSocketConnection) error     { return nil }
func (r *InMemoryRegistry) UpdateRegistry(c *WebSocketConnection) error { return nil }
func (r *InMemoryRegistry) Ping() error                                 { return nil }
func (r *InMemoryRegistry) Close() error                                { return nil }
