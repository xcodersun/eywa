package connections

import (
	"github.com/vivowares/eywa/Godeps/_workspace/src/github.com/satori/go.uuid"
	. "github.com/vivowares/eywa/Godeps/_workspace/src/github.com/smartystreets/goconvey/convey"
	"sync"
	"testing"
)

func TestHttpConnection(t *testing.T) {

	Convey("replacing a connection, will close the old one", t, func() {
		cm, _ := NewConnectionManager("default")
		defer CloseConnectionManager("default")

		ch1 := make(chan []byte, 1)
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
		httpConn, err := cm.NewHttpConnection("test", reqId1, ch1, h, nil)
		So(err, ShouldBeNil)
		wgConn.Wait()
		So(connected, ShouldBeTrue)

		//now register a new connection with the same id
		//the old one should be closed
		ch2 := make(chan []byte, 1)
		reqId2 := uuid.NewV4().String()
		_httpConn, err := cm.NewHttpConnection("test", reqId2, ch2, func(Connection, *Message, error) {}, nil)
		So(err, ShouldBeNil)
		So(cm.Count(), ShouldEqual, 1)

		wgDisconn.Wait()
		So(disconnected, ShouldBeTrue)
		So(httpConn.Closed(), ShouldBeTrue)
		So(_httpConn.Closed(), ShouldBeFalse)
	})

	Convey("sends message and closes an registered http connection", t, func() {
		cm, _ := NewConnectionManager("default")
		defer CloseConnectionManager("default")

		ch := make(chan []byte, 1)
		reqId := uuid.NewV4().String()
		_, err := cm.NewHttpConnection("test", reqId, ch, func(Connection, *Message, error) {}, nil)
		So(err, ShouldBeNil)

		httpConn, found := cm.FindConnection("test")
		So(found, ShouldBeTrue)
		err = httpConn.Send([]byte("message"))
		So(err, ShouldBeNil)

		msg := <-ch
		So(string(msg), ShouldEqual, "message")
		So(httpConn.Closed(), ShouldBeTrue)
		So(cm.Count(), ShouldEqual, 0)
	})
}
