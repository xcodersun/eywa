package connections

import (
	"fmt"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestWebsocketMessageSerializations(t *testing.T) {

	Convey("websocket message marshalling returns proper errors", t, func() {
		m := &websocketMessage{}
		_, err := m.Marshal()
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldEqual, "unsupported websocket message type 0")

		m._type = TypeConnectMessage
		_, err = m.Marshal()
		So(err, ShouldBeNil)
		So(len(m.payload), ShouldEqual, 0)
		So(len(m.id), ShouldBeGreaterThan, 0)
		So(len(m.raw), ShouldBeGreaterThan, len(m.payload))

		m = &websocketMessage{}
		m._type = TypeDisconnectMessage
		_, err = m.Marshal()
		So(err, ShouldBeNil)
		So(len(m.payload), ShouldEqual, 0)
		So(len(m.id), ShouldBeGreaterThan, 0)
		So(len(m.raw), ShouldBeGreaterThan, len(m.payload))

		m = &websocketMessage{}
		m._type = TypeRequestMessage
		_, err = m.Marshal()
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldEqual, "missing payload for websocket message type request")

		m = &websocketMessage{}
		m._type = TypeRequestMessage
		m.payload = []byte{}
		_, err = m.Marshal()
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldEqual, "missing message id for websocket message type request")

		m = &websocketMessage{}
		m._type = TypeRequestMessage
		m.payload = []byte{}
		m.id = "123"
		_, err = m.Marshal()
		So(err, ShouldBeNil)
		So(string(m.raw), ShouldEqual, fmt.Sprintf("%d|%s|", m._type, m.id))
	})

	Convey("websocket message unmarshalling returns proper errors", t, func() {
		m := &websocketMessage{}
		err := m.Unmarshal()
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldEqual, "missing message body for websocket message")

		m = &websocketMessage{}
		m._type = TypeConnectMessage
		err = m.Unmarshal()
		So(err, ShouldBeNil)
		So(m.payload, ShouldNotBeNil)
		So(m.raw, ShouldNotBeNil)
		So(len(m.id), ShouldBeGreaterThan, 0)

		m = &websocketMessage{}
		m._type = TypeSendMessage
		m.raw = []byte("")
		err = m.Unmarshal()
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldEqual, "expected 2 pips instead of 0 for websocket message type send")

		m = &websocketMessage{}
		m._type = TypeSendMessage
		m.raw = []byte("0|")
		err = m.Unmarshal()
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldEqual, "unsupported websocket message type 0")

		m = &websocketMessage{}
		m.raw = []byte(fmt.Sprintf("%d|", TypeSendMessage))
		err = m.Unmarshal()
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldEqual, "expected 2 pips instead of 1 for websocket message type send")

		m = &websocketMessage{}
		m.raw = []byte(fmt.Sprintf("%d|||", TypeRequestMessage))
		err = m.Unmarshal()
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldEqual, "empty message id for websocket message type request")

		m = &websocketMessage{}
		m.raw = []byte(fmt.Sprintf("%d|1|abc|def", TypeRequestMessage))
		err = m.Unmarshal()
		So(err, ShouldBeNil)
		So(m.id, ShouldEqual, "1")
		So(string(m.payload), ShouldEqual, "abc|def")
	})
}
