package connections

import (
	. "github.com/vivowares/eywa/Godeps/_workspace/src/github.com/smartystreets/goconvey/convey"
	. "github.com/vivowares/eywa/configs"
	. "github.com/vivowares/eywa/utils"
	"testing"
	"time"
)

func TestHttpConnection(t *testing.T) {

	SetConfig(&Conf{
		Connections: &ConnectionsConf{
			Registry:      "memory",
			NShards:       2,
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

	Convey("sends message and closes an registered http connection", t, func() {
		cm, _ := NewConnectionManager()
		defer cm.Close()

		ch := make(chan []byte, 1)
		_, err := cm.NewHttpConnection("test", ch, func(Connection, *Message, error) {}, nil)
		So(err, ShouldBeNil)

		httpConn, found := cm.FindConnection("test")
		So(found, ShouldBeTrue)
		err = httpConn.Send([]byte("message"))
		So(err, ShouldBeNil)

		msg := <-ch
		So(string(msg), ShouldEqual, "message")

		httpConn.Close()
		_, found = cm.FindConnection("test")
		So(found, ShouldBeFalse)

		err = httpConn.Send([]byte("another message"))
		So(err, ShouldNotBeNil)
	})
}
