package connections

import (
	"errors"
	"sync"
	"time"
)

var httpClosedErr = errors.New("http connection is closed")

type HttpConnection struct {
	requestId  string
	identifier string
	h          MessageHandler
	ch         chan []byte
	metadata   map[string]interface{}
	closed     bool
	createdAt  time.Time
	closedAt   time.Time
	closeOnce  sync.Once

	cm *ConnectionManager
}

func (c *HttpConnection) RequestId() string { return c.requestId }

func (c *HttpConnection) Identifier() string { return c.identifier }

func (c *HttpConnection) Metadata() map[string]interface{} { return c.metadata }

func (c *HttpConnection) MessageHandler() MessageHandler { return c.h }

func (c *HttpConnection) CreatedAt() time.Time { return c.createdAt }

func (c *HttpConnection) ClosedAt() time.Time { return c.closedAt }

func (c *HttpConnection) Closed() bool { return c.closed }

func (c *HttpConnection) LastPingedAt() time.Time { return c.createdAt }

func (c *HttpConnection) Send(msg []byte) (err error) {
	defer c.close(true)
	defer func() {
		if r := recover(); r != nil {
			err = httpClosedErr
		}
	}()

	c.ch <- msg
	return
}

func (c *HttpConnection) close(unregister bool) error {
	c.closeOnce.Do(func() {
		c.closed = true
		c.closedAt = time.Now()
		close(c.ch)
		if unregister {
			c.cm.unregister(c)
		}
		go c.h(c, &Message{MessageType: TypeDisconnectMessage}, nil)
	})
	return nil
}

func (c *HttpConnection) wait() {}
func (c *HttpConnection) ConnectionType() string {
	return "http"
}

func (c *HttpConnection) start() {
	go c.h(c, &Message{MessageType: TypeConnectMessage}, nil)
}
