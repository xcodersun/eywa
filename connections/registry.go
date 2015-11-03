package connections

type Registry interface {
	RegisterConnection(Connection) error
	UnregisterConnection(Connection) error
	FindConnection(string) (Connection, bool)
	ConnectionStats() *ConnectionStats
}

type ConnectionStats struct {
}
