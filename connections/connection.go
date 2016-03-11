package connections

import (
	"time"
)

type Connection interface {
	Identifier() string
	Close() error
	Closed() bool
	Wait()
	ConnectionType() string
	CreatedAt() time.Time
	ClosedAt() time.Time
	LastPingedAt() time.Time
	Metadata() map[string]interface{}
	MessageHandler() MessageHandler
	Send([]byte) error
	Start()
}
