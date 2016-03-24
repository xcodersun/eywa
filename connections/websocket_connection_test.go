package connections

import (
	"errors"
	"fmt"
	"github.com/vivowares/eywa/Godeps/_workspace/src/github.com/gorilla/websocket"
	"github.com/vivowares/eywa/Godeps/_workspace/src/github.com/satori/go.uuid"
	. "github.com/vivowares/eywa/Godeps/_workspace/src/github.com/smartystreets/goconvey/convey"
	. "github.com/vivowares/eywa/configs"
	. "github.com/vivowares/eywa/utils"
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
	syncSleepTime    time.Duration
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
	if strings.HasSuffix(m, "message sync") {
		msg, _ := Unmarshal(f.message)
		time.Sleep(f.syncSleepTime)
		return websocket.BinaryMessage, []byte(fmt.Sprintf("%d|%s|response sync", TypeResponseMessage, msg.MessageId)), nil
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

func TestWsConnection(t *testing.T) {

	SetConfig(&Conf{
		Connections: &ConnectionsConf{
			Websocket: &WsConnectionConf{
				RequestQueueSize: 8,
				Timeouts: &WsConnectionTimeoutConf{
					Write:    &JSONDuration{2 * time.Second},
					Read:     &JSONDuration{300 * time.Second},
					Request:  &JSONDuration{1 * time.Second},
					Response: &JSONDuration{2 * time.Second},
				},
				BufferSizes: &WsConnectionBufferSizeConf{
					Write: 1024,
					Read:  1024,
				},
			},
		},
	})

	h := func(c Connection, m *Message, e error) {}
	meta := make(map[string]interface{})

	Convey("errors out for request/response timeout", t, func() {
		conn := &WebsocketConnection{
			ws:         &fakeWsConn{},
			identifier: "test",
			h:          h,
			wch:        make(chan *MessageReq, 1),
		}

		err := conn.Response([]byte("resp"))
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldContainSubstring, "request timed out for 1s")

		err = conn.Send([]byte("async"))
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldContainSubstring, "request timed out for 1s")

		_, err = conn.Request([]byte("sync"), 3*time.Second)
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldContainSubstring, "request timed out for 1s")

		cm, err := NewConnectionManager("default")
		defer CloseConnectionManager("default")
		reqId := uuid.NewV4().String()
		conn, _ = cm.NewWebsocketConnection("test", reqId, &fakeWsConn{syncSleepTime: 5 * time.Second}, h, meta)
		So(cm.Count(), ShouldEqual, 1)

		_, err = conn.Request([]byte("sync"), 3*time.Second)
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldContainSubstring, "response timed out for 3s")
		So(conn.msgChans.len(), ShouldEqual, 0)
	})

	Convey("errors out send to closed connection", t, func() {
		cm, err := NewConnectionManager("default")
		defer CloseConnectionManager("default")

		reqId := uuid.NewV4().String()
		conn, _ := cm.NewWebsocketConnection("test", reqId, &fakeWsConn{}, h, meta)
		So(cm.Count(), ShouldEqual, 1)

		conn.close(true)
		conn.wait()
		So(cm.Count(), ShouldEqual, 0)
		err = conn.Send([]byte("message async"))
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldContainSubstring, "connection is closed")
	})

	Convey("closes connection after write/read error", t, func() {
		cm, err := NewConnectionManager("default")
		defer CloseConnectionManager("default")

		ws := &fakeWsConn{writeErr: errors.New("write err")}
		reqId := uuid.NewV4().String()
		conn, _ := cm.NewWebsocketConnection("test write err", reqId, ws, h, meta)
		So(cm.Count(), ShouldEqual, 1)

		err = conn.Send([]byte("async"))
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldContainSubstring, "WebsocketError")
		conn.wait()
		So(ws.closed, ShouldBeTrue)
		So(cm.Count(), ShouldEqual, 0)

		reqId = uuid.NewV4().String()
		ws = &fakeWsConn{readMessageErr: errors.New("read err")}
		conn, _ = cm.NewWebsocketConnection("test read err", reqId, ws, h, meta)

		conn.wait()
		So(ws.closed, ShouldBeTrue)
		So(cm.Count(), ShouldEqual, 0)
	})

	Convey("successfully sends async messages", t, func() {
		cm, err := NewConnectionManager("default")
		defer CloseConnectionManager("default")

		reqId := uuid.NewV4().String()
		fw := &fakeWsConn{}
		conn, _ := cm.NewWebsocketConnection("test", reqId, fw, h, meta)
		So(cm.Count(), ShouldEqual, 1)

		err = conn.Send([]byte("message async"))
		So(err, ShouldBeNil)
		So(strings.HasSuffix(string(fw.message), "message async"), ShouldBeTrue)
	})

	Convey("successfully sends sync messages", t, func() {
		cm, err := NewConnectionManager("default")
		defer CloseConnectionManager("default")

		reqId := uuid.NewV4().String()
		fw := &fakeWsConn{}
		conn, _ := cm.NewWebsocketConnection("test", reqId, fw, h, meta)
		So(cm.Count(), ShouldEqual, 1)

		msg, err := conn.Request([]byte("message sync"), Config().Connections.Websocket.Timeouts.Response.Duration)
		So(err, ShouldBeNil)
		So(string(msg), ShouldContainSubstring, "response sync")
		So(conn.msgChans.len(), ShouldEqual, 0)
	})

	Convey("replacing a connection, will close the old one", t, func() {
		cm, _ := NewConnectionManager("default")
		defer CloseConnectionManager("default")

		reqId1 := uuid.NewV4().String()
		var wgConn sync.WaitGroup
		var wgDisconn sync.WaitGroup
		connected := false
		disconnected := false
		wgConn.Add(1)    // expect to see connect message handled
		wgDisconn.Add(1) // expect to see disconnect message handled
		h := func(c Connection, m *Message, e error) {
			if m != nil {
				if m.MessageType == TypeConnectMessage {
					connected = true
					wgConn.Done()
				} else if m.MessageType == TypeDisconnectMessage {
					disconnected = true
					wgDisconn.Done()
				}
			}
		}

		httpConn, err := cm.NewWebsocketConnection("test", reqId1, &fakeWsConn{}, h, meta)
		So(err, ShouldBeNil)
		So(cm.Count(), ShouldEqual, 1)
		wgConn.Wait()
		So(connected, ShouldBeTrue)

		//now register a new connection with the same id
		//the old one should be closed
		reqId2 := uuid.NewV4().String()
		_httpConn, err := cm.NewWebsocketConnection("test", reqId2, &fakeWsConn{}, func(Connection, *Message, error) {}, meta)
		So(err, ShouldBeNil)
		So(cm.Count(), ShouldEqual, 1)

		wgDisconn.Wait()
		So(disconnected, ShouldBeTrue)
		So(httpConn.Closed(), ShouldBeTrue)
		So(_httpConn.Closed(), ShouldBeFalse)
	})

}
