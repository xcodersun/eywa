package connections

var CM ConnectionManager

type ConnectionManager interface {
	Close() error
	ConnectionCount() int
	Registry
}
