package connections

import (
	. "github.com/vivowares/eywa/Godeps/_workspace/src/github.com/smartystreets/goconvey/convey"
	. "github.com/vivowares/eywa/configs"
	. "github.com/vivowares/eywa/utils"
	"testing"
	"time"
)

func TestConnectionManager(t *testing.T) {

	SetConfig(&Conf{
		Connections: &ConnectionsConf{
			Registry:      "memory",
			NShards:       4,
			InitShardSize: 8,
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
		_, err := cm.NewHttpConnection("test http", ch, func(Connection, *Message, error) {}, nil)
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
		_, err = cm.NewHttpConnection("test http", ch, func(Connection, *Message, error) {}, nil)
		So(err, ShouldNotBeNil)
		So(err, ShouldEqual, closedCMErr)
		So(cm.Count(), ShouldEqual, 0)
		_, ok := <-ch
		So(ok, ShouldBeFalse)
	})
}
