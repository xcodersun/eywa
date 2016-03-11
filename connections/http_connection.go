package connections

import (
	"errors"
	"fmt"
	. "github.com/vivowares/eywa/loggers"
	"sync"
	"time"
)

var httpClosedErr = errors.New("http connection is closed")

type HttpConnection struct {
	identifier string
	h          MessageHandler
	ch         chan []byte
	metadata   map[string]interface{}
	closed     bool
	createdAt  time.Time
	closedAt   time.Time
	closeOnce  sync.Once

	shard *shard
}

func (c *HttpConnection) Identifier() string { return c.identifier }

func (c *HttpConnection) Metadata() map[string]interface{} { return c.metadata }

func (c *HttpConnection) MessageHandler() MessageHandler { return c.h }

func (c *HttpConnection) CreatedAt() time.Time { return c.createdAt }

func (c *HttpConnection) ClosedAt() time.Time { return c.closedAt }

func (c *HttpConnection) Closed() bool { return c.closed }

func (c *HttpConnection) LastPingedAt() time.Time { return c.createdAt }

func (c *HttpConnection) Send(msg []byte) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = httpClosedErr
		}
	}()

	c.ch <- msg
	return
}

func (c *HttpConnection) Close() error {
	c.closeOnce.Do(func() {
		c.closed = true
		c.closedAt = time.Now()
		if c.ch != nil {
			close(c.ch)
		}
		if c.shard != nil {
			c.shard.unregister(c)
		}
		Logger.Debug(fmt.Sprintf("http connection: %s closed", c.Identifier()))
		c.h(c, &Message{MessageType: TypeDisconnectMessage}, nil)
	})
	return nil
}

func (c *HttpConnection) Wait() {}
func (c *HttpConnection) ConnectionType() string {
	return "http"
}

func (c *HttpConnection) Start() {
	c.h(c, &Message{MessageType: TypeConnectMessage}, nil)
}
