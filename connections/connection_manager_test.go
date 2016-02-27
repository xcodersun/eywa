package connections

import (
	. "github.com/vivowares/eywa/Godeps/_workspace/src/github.com/smartystreets/goconvey/convey"
	. "github.com/vivowares/eywa/configs"
	"testing"
	"time"
)

func TestConnectionManager(t *testing.T) {

	SetConfig(&Conf{
		WebSocketConnections: &WsConnectionConf{
			Registry:         "memory",
			NShards:          4,
			InitShardSize:    8,
			RequestQueueSize: 8,
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

	Convey("creates/registers/finds new connections.", t, func() {
		wscm, _ := newWebSocketConnectionManager()
		defer wscm.close()
		conn, _ := wscm.newConnection("test", &fakeWsConn{}, h, meta) // this connection should be started and registered
		So(wscm.count(), ShouldEqual, 1)

		// the fake ReadMessage() always return empty string, which will still keep updating the
		// pingedAt timestamp
		t1 := conn.LastPingedAt()
		time.Sleep(200 * time.Millisecond)
		t2 := conn.LastPingedAt()
		So(t1.Equal(t2), ShouldBeFalse)

		_, found := wscm.findConnection("test")
		So(found, ShouldBeTrue)
	})

	Convey("disallows creating/registering new connections on closed CM.", t, func() {
		wscm, _ := newWebSocketConnectionManager()
		wscm.close()

		ws := &fakeWsConn{}
		_, err := wscm.newConnection("test", ws, h, meta)
		So(ws.closed, ShouldBeTrue)
		So(err, ShouldNotBeNil)
		So(wscm.count(), ShouldEqual, 0)
	})
}
