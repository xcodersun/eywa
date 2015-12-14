package connections

type Registry interface {
	Register(*Connection) error
	Unregister(*Connection) error
	UpdateRegistry(*Connection) error
	Ping() error
	Close() error
}

type InMemoryRegistry struct{}

func (r *InMemoryRegistry) Register(c *Connection) error       { return nil }
func (r *InMemoryRegistry) Unregister(c *Connection) error     { return nil }
func (r *InMemoryRegistry) UpdateRegistry(c *Connection) error { return nil }
func (r *InMemoryRegistry) Ping() error                        { return nil }
func (r *InMemoryRegistry) Close() error                       { return nil }
