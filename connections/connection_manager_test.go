package connections

import (
	. "github.com/smartystreets/goconvey/convey"
	. "github.com/eywa/configs"
	. "github.com/eywa/utils"
	"reflect"
	"strconv"
	"testing"
	"time"
)

func TestConnectionManager(t *testing.T) {

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

	Convey("creates/registers/finds new connections.", t, func() {
		cm, _ := NewConnectionManager("default")
		conn, _ := cm.NewWebsocketConnection("test ws", &fakeWsConn{}, h, meta) // this connection should be started and registered
		So(cm.Count(), ShouldEqual, 1)

		// the fake ReadMessage() always return empty string, which will still keep updating the
		// pingedAt timestamp
		t1 := conn.LastPingedAt()
		time.Sleep(200 * time.Millisecond)
		t2 := conn.LastPingedAt()
		So(t1.Equal(t2), ShouldBeFalse)

		_, found := cm.FindConnection("test ws")
		So(found, ShouldBeTrue)

		ch := make(chan []byte, 1)
		poll := &httpConn{
			_type: HttpPoll,
			ch:    ch,
			body:  []byte("poll message"),
		}
		_, err := cm.NewHttpConnection("test http", poll, func(Connection, Message, error) {}, nil)
		So(err, ShouldBeNil)

		httpConn, found := cm.FindConnection("test http")
		So(found, ShouldBeTrue)
		So(cm.Count(), ShouldEqual, 2)

		CloseConnectionManager("default")
		So(cm.Count(), ShouldEqual, 0)

		_, ok := <-ch
		So(ok, ShouldBeFalse)
		So(httpConn.Closed(), ShouldBeTrue)
		So(conn.Closed(), ShouldBeTrue)
	})

	Convey("disallows creating/registering new connections on closed CM.", t, func() {
		cm, _ := NewConnectionManager("default")
		CloseConnectionManager("default")

		ws := &fakeWsConn{}
		_, err := cm.NewWebsocketConnection("test ws", ws, h, meta)
		So(ws.closed, ShouldBeTrue)
		So(err, ShouldNotBeNil)
		So(cm.Count(), ShouldEqual, 0)

		ch := make(chan []byte, 1)
		poll := &httpConn{
			_type: HttpPoll,
			ch:    ch,
			body:  []byte("poll message"),
		}
		_, err = cm.NewHttpConnection("test http", poll, func(Connection, Message, error) {}, nil)
		So(err, ShouldNotBeNil)
		So(err, ShouldEqual, closedCMErr)
		So(cm.Count(), ShouldEqual, 0)
		_, ok := <-ch
		So(ok, ShouldBeFalse)
	})

	Convey("test scan connections", t, func() {
		cm, _ := NewConnectionManager("default")
		defer CloseConnectionManager("default")

		n := 5
		for i := 1; i <= n; i++ {
			cm.NewWebsocketConnection("conn"+strconv.Itoa(i), &fakeWsConn{}, func(Connection, Message, error) {}, nil)
		}
		So(cm.Count(), ShouldEqual, n)

		conns := cm.Scan("", 0)
		So(len(conns), ShouldEqual, 0)

		conns = cm.Scan("", 3)
		So(len(conns), ShouldEqual, 3)
		connIds := make([]string, len(conns))
		for i, conn := range conns {
			connIds[i] = conn.Identifier()
		}
		So(reflect.DeepEqual(connIds, []string{"conn1", "conn2", "conn3"}), ShouldBeTrue)

		conns = cm.Scan("", n+3)
		So(len(conns), ShouldEqual, n)
		connIds = make([]string, len(conns))
		for i, conn := range conns {
			connIds[i] = conn.Identifier()
		}
		So(reflect.DeepEqual(connIds, []string{"conn1", "conn2", "conn3", "conn4", "conn5"}), ShouldBeTrue)

		conns = cm.Scan("conn1", 0)
		So(len(conns), ShouldEqual, 0)

		conns = cm.Scan("conn1", 3)
		So(len(conns), ShouldEqual, 3)
		connIds = make([]string, len(conns))
		for i, conn := range conns {
			connIds[i] = conn.Identifier()
		}
		So(reflect.DeepEqual(connIds, []string{"conn2", "conn3", "conn4"}), ShouldBeTrue)

		conns = cm.Scan("conn1", n+3)
		So(len(conns), ShouldEqual, n-1)
		connIds = make([]string, len(conns))
		for i, conn := range conns {
			connIds[i] = conn.Identifier()
		}
		So(reflect.DeepEqual(connIds, []string{"conn2", "conn3", "conn4", "conn5"}), ShouldBeTrue)

		conns = cm.Scan("conn0", 0)
		So(len(conns), ShouldEqual, 0)

		conns = cm.Scan("conn0", 3)
		So(len(conns), ShouldEqual, 3)
		connIds = make([]string, len(conns))
		for i, conn := range conns {
			connIds[i] = conn.Identifier()
		}
		So(reflect.DeepEqual(connIds, []string{"conn1", "conn2", "conn3"}), ShouldBeTrue)

		conns = cm.Scan("conn0", n+3)
		So(len(conns), ShouldEqual, n)
		connIds = make([]string, len(conns))
		for i, conn := range conns {
			connIds[i] = conn.Identifier()
		}
		So(reflect.DeepEqual(connIds, []string{"conn1", "conn2", "conn3", "conn4", "conn5"}), ShouldBeTrue)

	})
}
