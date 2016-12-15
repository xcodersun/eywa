package connections

import (
	"github.com/google/btree"
	"strings"
	"time"
)

var SupportedConnectionTypes = []string{"websocket", "http"}

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
	unregister()
	wait()
}

type Sender interface {
	Send([]byte) error
}

type Requester interface {
	Request([]byte, time.Duration) ([]byte, error)
}

// dummy struct to help find real connection in the btree
type Lesser struct {
	id string
}

func (l *Lesser) Less(than btree.Item) bool {
	return strings.Compare(l.id, than.(Connection).Identifier()) < 0
}

func (l *Lesser) Identifier() string                    { return l.id }
func (l *Lesser) Closed() bool                          { return false }
func (l *Lesser) ConnectionType() string                { return "" }
func (l *Lesser) CreatedAt() time.Time                  { return time.Now() }
func (l *Lesser) ClosedAt() time.Time                   { return time.Now() }
func (l *Lesser) LastPingedAt() time.Time               { return time.Now() }
func (l *Lesser) Metadata() map[string]string           { return nil }
func (l *Lesser) ConnectionManager() *ConnectionManager { return nil }
func (l *Lesser) start()                                {}
func (l *Lesser) close(bool) error                      { return nil }
func (l *Lesser) unregister()                           {}
func (l *Lesser) wait()                                 {}
