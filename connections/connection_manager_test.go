package connections

import (
	. "github.com/smartystreets/goconvey/convey"
	. "github.com/vivowares/octopus/configs"
	"testing"
	"time"
)

func TestConnectionManager(t *testing.T) {

	Config = &Conf{
		Connections: &ConnectionConf{
			Registry:         "memory",
			NShards:          4,
			InitShardSize:    8,
			RequestQueueSize: 8,
			Expiry:           300 * time.Second,
			Timeouts: &ConnectionTimeoutConf{
				Write:    2 * time.Second,
				Read:     300 * time.Second,
				Request:  1 * time.Second,
				Response: 2 * time.Second,
			},
			BufferSizes: &ConnectionBufferSizeConf{
				Write: 1024,
				Read:  1024,
			},
		},
	}

	h := func(c *Connection, m *Message, e error) {}
	meta := make(map[string]interface{})

	Convey("creates/registers/finds new connections.", t, func() {
		cm, _ := NewConnectionManager()
		defer cm.Close()
		conn, _ := cm.NewConnection("test", &fakeWsConn{}, h, meta) // this connection should be started and registered
		So(cm.Count(), ShouldEqual, 1)

		// the fake ReadMessage() always return empty string, which will still keep updating the
		// pingedAt timestamp
		t1 := conn.LastPingedAt()
		time.Sleep(200 * time.Millisecond)
		t2 := conn.LastPingedAt()
		So(t1.Equal(t2), ShouldBeFalse)

		_, found := cm.FindConnection("test")
		So(found, ShouldBeTrue)
	})

	Convey("disallows creating/registering new connections on closed CM.", t, func() {
		cm, _ := NewConnectionManager()
		cm.Close()

		ws := &fakeWsConn{}
		_, err := cm.NewConnection("test", ws, h, meta)
		So(ws.closed, ShouldBeTrue)
		So(err, ShouldNotBeNil)
		So(cm.Count(), ShouldEqual, 0)
	})
}
