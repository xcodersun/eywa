// +build all connections

package connections

import (
	"errors"
	"fmt"
	. "github.com/smartystreets/goconvey/convey"
	. "github.com/vivowares/octopus/configs"
	"io"
	"net"
	"testing"
	"time"
)

type fakeNetConn struct {
	io.Reader
	io.Writer
}

func (c fakeNetConn) Close() error                       { return nil }
func (c fakeNetConn) LocalAddr() net.Addr                { return nil }
func (c fakeNetConn) RemoteAddr() net.Addr               { return nil }
func (c fakeNetConn) SetDeadline(t time.Time) error      { return nil }
func (c fakeNetConn) SetReadDeadline(t time.Time) error  { return nil }
func (c fakeNetConn) SetWriteDeadline(t time.Time) error { return nil }

type fakeWsConn struct {
	closed           bool
	closeErr         error
	writeErr         error
	writeDeadlineErr error
	readErr          error
	readDeadlineErr  error
	pingHandler      func(string) error
	readMessageType  int
	readMessageBuf   []byte
	readMessageErr   error
}

func (f *fakeWsConn) Subprotocol() string { return "" }
func (f *fakeWsConn) Close() error {
	f.closed = true
	return f.closeErr
}
func (f *fakeWsConn) LocalAddr() net.Addr                             { return nil }
func (f *fakeWsConn) RemoteAddr() net.Addr                            { return nil }
func (f *fakeWsConn) WriteControl(i int, b []byte, t time.Time) error { return nil }
func (f *fakeWsConn) NextWriter(i int) (io.WriteCloser, error)        { return nil, nil }
func (f *fakeWsConn) WriteMessage(int, []byte) error {
	return f.writeErr
}
func (f *fakeWsConn) SetWriteDeadline(t time.Time) error {
	return f.writeDeadlineErr
}
func (f *fakeWsConn) NextReader() (int, io.Reader, error) {
	return 0, nil, nil
}
func (f *fakeWsConn) ReadMessage() (int, []byte, error) {
	return f.readMessageType, f.readMessageBuf, f.readMessageErr
}
func (f *fakeWsConn) SetReadDeadline(t time.Time) error {
	return f.readDeadlineErr
}
func (f *fakeWsConn) SetReadLimit(i int64) {}
func (f *fakeWsConn) SetPingHandler(h func(string) error) {
	f.pingHandler = h
}
func (f *fakeWsConn) SetPongHandler(h func(string) error) {}
func (f *fakeWsConn) UnderlyingConn() net.Conn {
	return &fakeNetConn{}
}

func TestConnection(t *testing.T) {

	Config = &Conf{
		Connections: &ConnectionConf{
			Store:  "memory",
			Expiry: 300 * time.Second,
			Timeouts: &ConnectionTimeoutConf{
				Write:    2 * time.Second,
				Read:     300 * time.Second,
				Response: 8 * time.Second,
			},
			BufferSizes: &ConnectionBufferSizeConf{
				Write: 1024,
				Read:  1024,
			},
		},
	}

	Convey("checks/sends async request and response.", t, func() {
		InitializeCM()

		var h = func(c Connection, m *Message, e error) {
		}
		ws := &fakeWsConn{}
		conn, err := CM.NewConnection("test", ws, h)
		So(err, ShouldBeNil)

		msg := &Message{
			MessageType: 0,
			MessageId:   "",
			Payload:     "",
		}
		err = conn.SendAsyncRequest(msg)
		err, ok := err.(*MessageSendingError)
		So(ok, ShouldBeTrue)
		So(err.Error(), ShouldContainSubstring, fmt.Sprintf("invalid message type %d, expected %d", 0, AsyncRequestMessage))

		msg.MessageType = AsyncRequestMessage
		err = conn.SendAsyncRequest(msg)
		err, ok = err.(*MessageIdError)
		So(ok, ShouldBeTrue)
		So(err.Error(), ShouldContainSubstring, "empty message id")

		msg.MessageId = "test id"
		ws.writeDeadlineErr = errors.New("test error")
		err = conn.SendAsyncRequest(msg)
		err, ok = err.(*WebsocketError)
		So(ok, ShouldBeTrue)
		So(err.Error(), ShouldContainSubstring, "error setting write deadline")

		ws.writeDeadlineErr = nil
		ws.writeErr = errors.New("test write error")
		err = conn.SendAsyncRequest(msg)
		err, ok = err.(*WebsocketError)
		So(ok, ShouldBeTrue)
		So(err.Error(), ShouldContainSubstring, "test write error")

		msg.MessageType = 0
		err = conn.SendResponse(msg)
		err, ok = err.(*MessageSendingError)
		So(ok, ShouldBeTrue)
		So(err.Error(), ShouldContainSubstring, fmt.Sprintf("invalid message type %d, expected %d", 0, ResponseMessage))

		conn.Close()
		msg.MessageType = AsyncRequestMessage
		err = conn.SendAsyncRequest(msg)
		err, ok = err.(*MessageSendingError)
		So(ok, ShouldBeTrue)
		So(err.Error(), ShouldContainSubstring, "connection closed")

		CM.Close()
		CM.Wait()
	})

	Convey("checks/sends sync request.", t, func() {
		InitializeCM()

		var h = func(c Connection, m *Message, e error) {
		}
		ws := &fakeWsConn{
			readMessageBuf: []byte(fmt.Sprintf("%d|%s|response", ResponseMessage, "test_sync")),
		}
		conn, err := CM.NewConnection("test", ws, h)
		So(err, ShouldBeNil)

		msg := &Message{
			MessageType: SyncRequestMessage,
			MessageId:   "test_sync",
			Payload:     "request",
		}
		resp, err := conn.SendSyncRequest(msg)
		So(err, ShouldBeNil)
		So(resp.Payload, ShouldEqual, "response")

		CM.Close()
		CM.Wait()
	})
}
