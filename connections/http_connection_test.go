package connections

import (
	. "github.com/smartystreets/goconvey/convey"
	"reflect"
	"sync"
	"testing"
	"time"
)

func TestHttpConnection(t *testing.T) {

	Convey("replacing an http connection, will close the old one", t, func() {
		cm, _ := NewConnectionManager("default")
		defer CloseConnectionManager("default")

		httpConn1 := &httpConn{
			_type: HttpPoll,
			body:  []byte{},
		}

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
		conn1, err := cm.NewHttpConnection("test", httpConn1, h, nil)
		So(err, ShouldBeNil)
		So(cm.Count(), ShouldEqual, 1)
		wgConn.Wait()
		So(connected, ShouldBeTrue)

		//now register a new connection with the same id
		//the old one should be closed
		httpConn2 := &httpConn{
			_type: HttpPoll,
			body:  []byte{},
		}
		conn2, err := cm.NewHttpConnection("test", httpConn2, func(Connection, Message, error) {}, nil)
		So(err, ShouldBeNil)
		So(cm.Count(), ShouldEqual, 1)

		wgDisconn.Wait()
		So(disconnected, ShouldBeTrue)
		So(conn1.Closed(), ShouldBeTrue)
		So(conn2.Closed(), ShouldBeFalse)
	})

	Convey("http push/poll messages' upload/send are handled", t, func() {
		cm, _ := NewConnectionManager("default")
		defer CloseConnectionManager("default")

		push := &httpConn{
			_type: HttpPush,
			body:  []byte("push message"),
		}
		var pushed, uploaded, sent, connected, disconnected bool
		var pushWg, uploadWg, sentWg, connWg, disConnWg sync.WaitGroup
		pushWg.Add(1)
		uploadWg.Add(1)
		sentWg.Add(1)
		connWg.Add(1)
		disConnWg.Add(1)
		h := func(c Connection, m Message, e error) {
			if m != nil {
				if m.Type() == TypeConnectMessage {
					connected = true
					connWg.Done()
				} else if m.Type() == TypeDisconnectMessage {
					disconnected = true
					disConnWg.Done()
				} else if m.Type() == TypeUploadMessage {
					if c.ConnectionType() == HttpConnectionTypes[HttpPush] {
						pushed = true
						pushWg.Done()
					} else {
						uploaded = true
						uploadWg.Done()
					}
				} else if m.Type() == TypeSendMessage {
					sent = true
					sentWg.Done()
				}
			}
		}
		pushConn, err := cm.NewHttpConnection("test", push, h, nil)
		So(err, ShouldBeNil)
		So(cm.Count(), ShouldEqual, 0)
		pushWg.Wait()
		So(pushed, ShouldBeTrue)
		So(uploaded, ShouldBeFalse)
		So(sent, ShouldBeFalse)
		So(connected, ShouldBeFalse)
		So(disconnected, ShouldBeFalse)
		So(pushConn.Closed(), ShouldBeTrue)

		poll := &httpConn{
			_type: HttpPoll,
			ch:    make(chan []byte, 1),
			body:  []byte("poll message"),
		}
		pollConn, err := cm.NewHttpConnection("test", poll, h, nil)
		So(err, ShouldBeNil)
		So(cm.Count(), ShouldEqual, 1)
		uploadWg.Wait()
		connWg.Wait()
		So(uploaded, ShouldBeTrue)
		So(sent, ShouldBeFalse)
		So(connected, ShouldBeTrue)
		So(disconnected, ShouldBeFalse)
		So(pollConn.Closed(), ShouldBeFalse)

		pollConn.Send([]byte("send message"))
		sentWg.Wait()
		So(sent, ShouldBeTrue)
		disConnWg.Wait()
		So(disconnected, ShouldBeTrue)
		So(pollConn.Closed(), ShouldBeTrue)
		So(cm.Count(), ShouldEqual, 0)
	})

	Convey("polls message with timeout", t, func() {
		cm, _ := NewConnectionManager("default")
		defer CloseConnectionManager("default")

		poll := &httpConn{
			_type: HttpPoll,
			ch:    make(chan []byte, 1),
			body:  []byte("poll message"),
		}

		pollConn, err := cm.NewHttpConnection("test", poll, func(Connection, Message, error) {}, nil)
		So(err, ShouldBeNil)
		So(cm.Count(), ShouldEqual, 1)

		var sentWg sync.WaitGroup
		sentWg.Add(1)
		go func() {
			time.Sleep(1 * time.Second)
			err = pollConn.Send([]byte("send message"))
			sentWg.Done()
		}()
		p := pollConn.Poll(5 * time.Millisecond)
		So(p, ShouldBeNil)
		So(pollConn.Closed(), ShouldBeTrue)
		So(cm.Count(), ShouldEqual, 0)
		sentWg.Wait()
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldEqual, "http poll connection is closed")
	})

	Convey("polls message without timeout", t, func() {
		cm, _ := NewConnectionManager("default")
		defer CloseConnectionManager("default")

		poll := &httpConn{
			_type: HttpPoll,
			ch:    make(chan []byte, 1),
			body:  []byte("poll message"),
		}

		pollConn, err := cm.NewHttpConnection("test", poll, func(Connection, Message, error) {}, nil)
		So(err, ShouldBeNil)
		So(cm.Count(), ShouldEqual, 1)

		msg := []byte("send message")
		var sentWg sync.WaitGroup
		sentWg.Add(1)
		go func() {
			err = pollConn.Send(msg)
			sentWg.Done()
		}()
		p := pollConn.Poll(1 * time.Second)
		So(p, ShouldNotBeNil)
		So(reflect.DeepEqual(p, msg), ShouldBeTrue)
		So(pollConn.Closed(), ShouldBeTrue)
		So(cm.Count(), ShouldEqual, 0)
		sentWg.Wait()
		So(err, ShouldBeNil)
	})
}
