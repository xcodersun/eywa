package connections

import (
	"time"
)

type Connection interface {
	RequestId() string
	Identifier() string
	Closed() bool
	ConnectionType() string
	CreatedAt() time.Time
	ClosedAt() time.Time
	LastPingedAt() time.Time
	Metadata() map[string]interface{}
	MessageHandler() MessageHandler
	Send([]byte) error

	start()
	close(bool) error
	wait()
}
