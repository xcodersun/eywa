package connections

import (
	"errors"
	"github.com/google/btree"
	"github.com/eywa/pubsub"
	"strings"
	"sync"
	"time"
)

var HttpCloseChan = make(chan struct{})

type HttpConnectionType uint8

const (
	HttpPush HttpConnectionType = iota
	HttpPoll
)

var HttpConnectionTypes = map[HttpConnectionType]string{
	HttpPush: "http push",
	HttpPoll: "http poll",
}

var httpPollClosedErr = errors.New("http poll connection is closed")

type httpConn struct {
	_type HttpConnectionType
	ch    chan []byte
	body  []byte
}

func (c *httpConn) read() []byte {
	return c.body
}

func (c *httpConn) write(p []byte) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = httpPollClosedErr
		}
	}()

	c.ch <- p
	return
}

func (c *httpConn) close() {
	if c.ch != nil {
		close(c.ch)
	}
}

type HttpConnection struct {
	identifier string
	h          MessageHandler
	httpConn   *httpConn
	metadata   map[string]string
	closed     bool
	createdAt  time.Time
	closedAt   time.Time
	closeOnce  sync.Once
	*pubsub.BasicPublisher

	cm *ConnectionManager
}

func (c *HttpConnection) Identifier() string { return c.identifier }

func (c *HttpConnection) Metadata() map[string]string { return c.metadata }

func (c *HttpConnection) CreatedAt() time.Time { return c.createdAt }

func (c *HttpConnection) ClosedAt() time.Time { return c.closedAt }

func (c *HttpConnection) Closed() bool { return c.closed }

func (c *HttpConnection) LastPingedAt() time.Time { return c.createdAt }

func (c *HttpConnection) ConnectionManager() *ConnectionManager { return c.cm }

func (c *HttpConnection) Less(than btree.Item) bool {
	conn := than.(Connection)
	return strings.Compare(c.identifier, conn.Identifier()) < 0
}

func (c *HttpConnection) Poll(dur time.Duration) []byte {
	defer c.close(true)

	select {
	case <-HttpCloseChan:
		return nil
	case <-time.After(dur):
		return nil
	case p, ok := <-c.httpConn.ch:
		if ok {
			return p
		} else {
			return nil
		}
	}
}

func (c *HttpConnection) Send(msg []byte) error {
	if c.httpConn._type != HttpPoll {
		return errors.New("only http poll connection supports message sending")
	}
	defer c.close(true)

	m := &httpMessage{_type: TypeSendMessage, raw: msg}
	p, err := m.Marshal()

	if err == nil {
		err = c.httpConn.write(p)
	}
	go c.h(c, m, err)
	return err
}

func (c *HttpConnection) unregister() {
	// To avoid race condition where a new connection has registered
	// under the same id and current connection become orphan, in which
	// case the orphan connection has different creatd time with the
	// registered connection
	conn, found := c.cm.FindConnection(c.identifier)
	if found && conn.CreatedAt() == c.createdAt {
		c.cm.unregister(c)
	}
}

func (c *HttpConnection) close(unregister bool) error {
	c.closeOnce.Do(func() {
		c.closed = true
		c.closedAt = time.Now()
		c.httpConn.close()
		if unregister && c.httpConn._type == HttpPoll {
			c.unregister()
		}
		if c.httpConn._type == HttpPoll {
			go c.h(c, &httpMessage{_type: TypeDisconnectMessage}, nil)
		}

		go func() {
			time.Sleep(3 * time.Second) // for user experience
			c.BasicPublisher.Unpublish()
		}()
	})
	return nil
}

func (c *HttpConnection) wait() {}
func (c *HttpConnection) ConnectionType() string {
	return HttpConnectionTypes[c.httpConn._type]
}

func (c *HttpConnection) start() {
	if c.httpConn._type == HttpPoll {
		go c.h(c, &httpMessage{_type: TypeConnectMessage}, nil)
	}

	go func() {
		m := &httpMessage{_type: TypeUploadMessage, raw: c.httpConn.read()}
		c.h(c, m, m.Unmarshal())
	}()
}
