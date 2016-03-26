package connections

import (
	"time"
)

type Connection interface {
	Identifier() string
	Closed() bool
	ConnectionType() string
	CreatedAt() time.Time
	ClosedAt() time.Time
	LastPingedAt() time.Time
	Metadata() map[string]string
	ConnectionManager() *ConnectionManager

	start()
	close(bool) error
	wait()
}

type Sender interface {
	Send([]byte) error
}

type Requester interface {
	Request([]byte, time.Duration) ([]byte, error)
}
