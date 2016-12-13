package connections

import (
	. "github.com/smartystreets/goconvey/convey"
	"reflect"
	"testing"
)

func TestHttpMessageSerializations(t *testing.T) {
	Convey("http message can be marshaled/unmarshaled", t, func() {
		raw := []byte("\u2318\u2317")
		m := &httpMessage{
			_type: TypeSendMessage,
			raw:   raw,
		}

		p, err := m.Marshal()
		So(err, ShouldBeNil)
		So(reflect.DeepEqual(raw, p), ShouldBeTrue)
		So(len(m.id), ShouldBeGreaterThan, 0)

		n := &httpMessage{_type: TypeSendMessage, raw: raw}
		err = n.Unmarshal()
		So(err, ShouldBeNil)
		So(reflect.DeepEqual(raw, n.raw), ShouldBeTrue)
		So(len(n.id), ShouldBeGreaterThan, 0)
	})

	Convey("http message marshalling returns proper errors", t, func() {
		m := &httpMessage{}
		_, err := m.Marshal()
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldEqual, "unsupported http message type 0")

		m._type = TypeSendMessage
		_, err = m.Marshal()
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldEqual, "missing message body for http message type send")

		m._type = TypeUploadMessage
		_, err = m.Marshal()
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldEqual, "missing message body for http message type upload")

		m._type = TypeConnectMessage
		_, err = m.Marshal()
		So(err, ShouldBeNil)
		So(len(m.id), ShouldBeGreaterThan, 0)

		n := &httpMessage{}
		n._type = TypeDisconnectMessage
		_, err = n.Marshal()
		So(err, ShouldBeNil)
		So(len(n.id), ShouldBeGreaterThan, 0)
	})

	Convey("http message unmarshalling returns proper errors", t, func() {
		m := &httpMessage{}
		err := m.Unmarshal()
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldEqual, "unsupported http message type 0")

		m._type = TypeSendMessage
		err = m.Unmarshal()
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldEqual, "missing message body for http message type send")

		m._type = TypeUploadMessage
		err = m.Unmarshal()
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldEqual, "missing message body for http message type upload")

		m._type = TypeConnectMessage
		err = m.Unmarshal()
		So(err, ShouldBeNil)
		So(len(m.id), ShouldBeGreaterThan, 0)

		n := &httpMessage{}
		n._type = TypeDisconnectMessage
		err = n.Unmarshal()
		So(err, ShouldBeNil)
		So(len(n.id), ShouldBeGreaterThan, 0)
	})
}
