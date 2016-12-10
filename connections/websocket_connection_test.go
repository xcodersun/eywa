package connections

import (
	"errors"
	"fmt"
	"github.com/gorilla/websocket"
	. "github.com/smartystreets/goconvey/convey"
	. "github.com/eywa/configs"
	. "github.com/eywa/utils"
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
	requested        *sync.WaitGroup
	responsed        bool
	uploaded         bool
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
	if strings.HasSuffix(m, "message sync") && !f.responsed {
		f.responsed = true
		msg := &websocketMessage{raw: []byte(m)}
		msg.Unmarshal()
		time.Sleep(f.syncSleepTime)
		if f.requested != nil {
			f.requested.Wait()
		}
		return websocket.BinaryMessage, []byte(fmt.Sprintf("%d|%s|response sync", TypeResponseMessage, msg.id)), nil
	}

	if !f.uploaded {
		f.uploaded = true
		return websocket.BinaryMessage, []byte(fmt.Sprintf("%d|%s|message upload", TypeUploadMessage, "random id")), nil
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

	h := func(c Connection, m Message, e error) {}
	meta := make(map[string]string)

	Convey("errors out for request/response timeout", t, func() {
		conn := &WebsocketConnection{
			ws:         &fakeWsConn{},
			identifier: "test",
			h:          h,
			wch:        make(chan *websocketMessageReq, 1),
		}

		err := conn.Send([]byte("async"))
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldContainSubstring, "request timed out for 1s")

		_, err = conn.Request([]byte("sync"), 3*time.Second)
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldContainSubstring, "request timed out for 1s")

		cm, err := NewConnectionManager("default")
		defer CloseConnectionManager("default")
		conn, _ = cm.NewWebsocketConnection("test", &fakeWsConn{syncSleepTime: 5 * time.Second}, h, meta)
		So(cm.Count(), ShouldEqual, 1)

		_, err = conn.Request([]byte("sync"), 3*time.Second)
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldContainSubstring, "response timed out for 3s")
		So(conn.msgChans.len(), ShouldEqual, 0)
	})

	Convey("errors out send to closed connection", t, func() {
		cm, err := NewConnectionManager("default")
		defer CloseConnectionManager("default")

		conn, _ := cm.NewWebsocketConnection("test", &fakeWsConn{}, h, meta)
		So(cm.Count(), ShouldEqual, 1)

		conn.close(true)
		conn.wait()
		So(cm.Count(), ShouldEqual, 0)
		err = conn.Send([]byte("message async"))
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldContainSubstring, "websocket connection is closed")
	})

	Convey("closes connection after write/read error", t, func() {
		cm, err := NewConnectionManager("default")
		defer CloseConnectionManager("default")

		ws := &fakeWsConn{writeErr: errors.New("write err")}
		conn, _ := cm.NewWebsocketConnection("test write err", ws, h, meta)
		So(cm.Count(), ShouldEqual, 1)

		err = conn.Send([]byte("async"))
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldContainSubstring, "WebsocketError")
		conn.wait()
		So(ws.closed, ShouldBeTrue)
		So(cm.Count(), ShouldEqual, 0)

		ws = &fakeWsConn{readMessageErr: errors.New("read err")}
		conn, _ = cm.NewWebsocketConnection("test read err", ws, h, meta)
		conn.wait()
		So(ws.closed, ShouldBeTrue)
		So(cm.Count(), ShouldEqual, 0)
	})

	Convey("successfully sends async messages", t, func() {
		cm, err := NewConnectionManager("default")
		defer CloseConnectionManager("default")

		fw := &fakeWsConn{}
		conn, _ := cm.NewWebsocketConnection("test", fw, h, meta)
		So(cm.Count(), ShouldEqual, 1)

		err = conn.Send([]byte("message async"))
		So(err, ShouldBeNil)
		So(strings.HasSuffix(string(fw.message), "message async"), ShouldBeTrue)
	})

	Convey("successfully sends sync messages", t, func() {
		cm, err := NewConnectionManager("default")
		defer CloseConnectionManager("default")

		var reqWg sync.WaitGroup
		reqWg.Add(1)
		h := func(c Connection, m Message, err error) {
			if m != nil && m.Type() == TypeRequestMessage {
				reqWg.Done()
			}
		}
		fw := &fakeWsConn{requested: &reqWg}
		conn, _ := cm.NewWebsocketConnection("test", fw, h, meta)
		So(cm.Count(), ShouldEqual, 1)

		msg, err := conn.Request([]byte("message sync"), Config().Connections.Websocket.Timeouts.Response.Duration)
		So(err, ShouldBeNil)
		So(string(msg), ShouldContainSubstring, "response sync")
		So(conn.msgChans.len(), ShouldEqual, 0)
	})

	Convey("replacing a connection, will close the old one", t, func() {
		cm, _ := NewConnectionManager("default")
		defer CloseConnectionManager("default")

		var wgConn sync.WaitGroup
		var wgDisconn sync.WaitGroup
		connected := false
		disconnected := false
		wgConn.Add(1)    // expect to see connect message handled
		wgDisconn.Add(1) // expect to see disconnect message handled
		h := func(c Connection, m Message, e error) {
			if m != nil {
				if m.Type() == TypeConnectMessage {
					connected = true
					wgConn.Done()
				} else if m.Type() == TypeDisconnectMessage {
					disconnected = true
					wgDisconn.Done()
				}
			}
		}

		httpConn, err := cm.NewWebsocketConnection("test", &fakeWsConn{}, h, meta)
		So(err, ShouldBeNil)
		So(cm.Count(), ShouldEqual, 1)
		wgConn.Wait()
		So(connected, ShouldBeTrue)
		So(disconnected, ShouldBeFalse)

		//now register a new connection with the same id
		//the old one should be closed
		_httpConn, err := cm.NewWebsocketConnection("test", &fakeWsConn{}, func(Connection, Message, error) {}, meta)
		So(err, ShouldBeNil)
		So(cm.Count(), ShouldEqual, 1)

		wgDisconn.Wait()
		So(disconnected, ShouldBeTrue)
		So(httpConn.Closed(), ShouldBeTrue)
		So(_httpConn.Closed(), ShouldBeFalse)
	})

	Convey("captures the messages in handler during the connection life cycle", t, func() {
		cm, _ := NewConnectionManager("default")
		defer CloseConnectionManager("default")

		var connected, disconnected, sent, uploaded, requested, responsed bool
		var connWg, disConnWg, sentWg, uploadWg, requestWg, responseWg sync.WaitGroup
		connWg.Add(1)
		disConnWg.Add(1)
		sentWg.Add(1)
		uploadWg.Add(1)
		requestWg.Add(1)
		responseWg.Add(1)

		h := func(c Connection, m Message, e error) {
			if m != nil {
				if m.Type() == TypeConnectMessage {
					connected = true
					connWg.Done()
				} else if m.Type() == TypeDisconnectMessage {
					disconnected = true
					disConnWg.Done()
				} else if m.Type() == TypeSendMessage {
					sent = true
					sentWg.Done()
				} else if m.Type() == TypeUploadMessage {
					uploaded = true
					uploadWg.Done()
				} else if m.Type() == TypeRequestMessage {
					requested = true
					requestWg.Done()
				} else if m.Type() == TypeResponseMessage {
					responsed = true
					responseWg.Done()
				}
			}
		}

		conn, err := cm.NewWebsocketConnection("test", &fakeWsConn{requested: &requestWg}, h, meta)
		So(err, ShouldBeNil)
		So(cm.Count(), ShouldEqual, 1)
		connWg.Wait()
		uploadWg.Wait()
		So(connected, ShouldBeTrue)
		So(uploaded, ShouldBeTrue)
		So(sent, ShouldBeFalse)
		So(requested, ShouldBeFalse)
		So(responsed, ShouldBeFalse)
		So(disconnected, ShouldBeFalse)

		err = conn.Send([]byte("message async"))
		So(err, ShouldBeNil)
		sentWg.Wait()
		So(sent, ShouldBeTrue)
		So(requested, ShouldBeFalse)
		So(responsed, ShouldBeFalse)
		So(disconnected, ShouldBeFalse)

		_, err = conn.Request([]byte("message sync"), Config().Connections.Websocket.Timeouts.Response.Duration)
		So(err, ShouldBeNil)
		requestWg.Wait()
		responseWg.Wait()
		So(requested, ShouldBeTrue)
		So(responsed, ShouldBeTrue)
		So(disconnected, ShouldBeFalse)
		So(conn.msgChans.len(), ShouldEqual, 0)

		conn.close(true)
		disConnWg.Wait()
		So(cm.Count(), ShouldEqual, 0)
		So(disconnected, ShouldBeTrue)
	})
}
