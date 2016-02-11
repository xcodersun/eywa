package connections

import (
	"errors"
	"fmt"
	"github.com/vivowares/octopus/Godeps/_workspace/src/github.com/gorilla/websocket"
	. "github.com/vivowares/octopus/Godeps/_workspace/src/github.com/smartystreets/goconvey/convey"
	. "github.com/vivowares/octopus/configs"
	"io"
	"math/rand"
	"net"
	"strings"
	"sync"
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
	readDeadlineErr  error
	pingHandler      func(string) error
	readMessageType  int
	readMessageBuf   []byte
	readMessageErr   error
	randomErr        bool
	message          []byte
	sync.Mutex
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
func (f *fakeWsConn) WriteMessage(msgType int, msg []byte) error {
	f.Lock()
	f.message = msg
	f.Unlock()
	time.Sleep(time.Duration(rand.Intn(200)) * time.Millisecond)
	if f.randomErr && rand.Intn(3) == 0 {
		f.writeErr = errors.New("write err")
	}
	return f.writeErr
}
func (f *fakeWsConn) SetWriteDeadline(t time.Time) error {
	return f.writeDeadlineErr
}
func (f *fakeWsConn) NextReader() (int, io.Reader, error) {
	return 0, nil, nil
}
func (f *fakeWsConn) ReadMessage() (int, []byte, error) {
	time.Sleep(time.Duration(rand.Intn(200)) * time.Millisecond)
	f.Lock()
	m := string(f.message)
	f.Unlock()
	if strings.HasSuffix(m, "sync") {
		msg, _ := Unmarshal(f.message)
		return websocket.BinaryMessage, []byte(fmt.Sprintf("%d|%s|sync response", ResponseMessage, msg.MessageId)), nil
	}

	if f.randomErr && rand.Intn(3) == 0 {
		f.readMessageErr = errors.New("read err")
	}
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

func TestConnections(t *testing.T) {

	SetConfig(&Conf{
		WebSocketConnections: &WsConnectionConf{
			Registry:         "memory",
			NShards:          2,
			InitShardSize:    8,
			RequestQueueSize: 8,
			Expiry:           300 * time.Second,
			Timeouts: &WsConnectionTimeoutConf{
				Write:    2 * time.Second,
				Read:     300 * time.Second,
				Request:  1 * time.Second,
				Response: 2 * time.Second,
			},
			BufferSizes: &WsConnectionBufferSizeConf{
				Write: 1024,
				Read:  1024,
			},
		},
	})

	h := func(c Connection, m *Message, e error) {}
	meta := make(map[string]interface{})

	Convey("errors out for request/response timeout", t, func() {
		conn := &WebSocketConnection{
			ws:         &fakeWsConn{},
			identifier: "test",
			h:          h,
			wch:        make(chan *MessageReq, 1),
		}

		err := conn.SendResponse([]byte("resp"))
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldContainSubstring, "response timed out")

		err = conn.SendAsyncRequest([]byte("async"))
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldContainSubstring, "request timed out")
	})

	Convey("errors out closed connection", t, func() {
		InitializeWSCM();
		defer CloseWSCM()
		conn, _ := NewWebSocketConnection("test", &fakeWsConn{}, h, meta)
		So(WebSocketCount(), ShouldEqual, 1)

		conn.Close()
		conn.Wait()
		So(WebSocketCount(), ShouldEqual, 0)
		err := conn.SendAsyncRequest([]byte("async"))
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldContainSubstring, "connection is closed")
	})

	Convey("closes connection after write/read error", t, func() {
		InitializeWSCM()
		defer CloseWSCM()
		ws := &fakeWsConn{writeErr: errors.New("write err")}
		conn, _ := NewWebSocketConnection("test write err", ws, h, meta)
		So(WebSocketCount(), ShouldEqual, 1)

		err := conn.SendAsyncRequest([]byte("async"))
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldContainSubstring, "WebsocketError")
		conn.Wait()
		So(ws.closed, ShouldBeTrue)
		So(WebSocketCount(), ShouldEqual, 0)

		ws = &fakeWsConn{readMessageErr: errors.New("read err")}
		conn, _ = NewWebSocketConnection("test read err", ws, h, meta)

		conn.Wait()
		So(ws.closed, ShouldBeTrue)
		So(WebSocketCount(), ShouldEqual, 0)
	})

	Convey("successfully sends async messages", t, func() {
		InitializeWSCM()
		defer CloseWSCM()
		conn, _ := NewWebSocketConnection("test", &fakeWsConn{}, h, meta)
		So(WebSocketCount(), ShouldEqual, 1)

		err := conn.SendAsyncRequest([]byte("async"))
		So(err, ShouldBeNil)
	})

	Convey("successfully sends sync messages", t, func() {
		InitializeWSCM()
		defer CloseWSCM()
		conn, _ := NewWebSocketConnection("test", &fakeWsConn{}, h, meta)
		So(WebSocketCount(), ShouldEqual, 1)

		msg, err := conn.SendSyncRequest([]byte("sync"))
		So(err, ShouldBeNil)
		So(string(msg), ShouldContainSubstring, "sync response")
		So(conn.msgChans.len(), ShouldEqual, 0)
	})

}
