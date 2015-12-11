// +build all connections

package connections

import (
	. "github.com/smartystreets/goconvey/convey"
	. "github.com/vivowares/octopus/configs"
	"strconv"
	"testing"
	"time"
)

func TestConnectionManager(t *testing.T) {

	Config = &Conf{
		Connections: &ConnectionConf{
			Store:  "memory",
			Expiry: 1 * time.Second,
			Timeouts: &ConnectionTimeoutConf{
				Write:    2 * time.Second,
				Read:     1 * time.Second,
				Response: 8 * time.Second,
			},
			BufferSizes: &ConnectionBufferSizeConf{
				Write: 1024,
				Read:  1024,
			},
		},
	}

	var h = func(c Connection, m *Message, e error) {
	}

	Convey("creates/registers new connections.", t, func() {
		InitializeCM()

		ws := &fakeWsConn{}
		conn, err := CM.NewConnection("test", ws, h)
		So(err, ShouldBeNil)
		So(CM.ConnectionCount(), ShouldEqual, 1)

		// the fake ReadMessage() always return empty string, which will still keep updating the
		// pingedAt timestamp
		t1 := conn.LastPingedAt()
		time.Sleep(1 * time.Second)
		t2 := conn.LastPingedAt()
		So(t1.Equal(t2), ShouldBeFalse)
		CM.Close()
		CM.Wait()
	})

	Convey("disallows creating/registering new connections on closed CM.", t, func() {
		InitializeCM()
		CM.Close()

		ws := &fakeWsConn{
			closed: false,
		}
		_, err := CM.NewConnection("test", ws, h)
		So(err, ShouldNotBeNil)
		So(CM.ConnectionCount(), ShouldEqual, 0)
		So(ws.closed, ShouldBeTrue)
	})

	Convey("won't unregister the connection if the connection got reconnected.", t, func() {
		InitializeCM()

		ws1 := &fakeWsConn{closed: false}
		conn1, err := CM.NewConnection("test", ws1, h)
		So(err, ShouldBeNil)
		So(CM.ConnectionCount(), ShouldEqual, 1)

		time.Sleep(1 * time.Second)
		ws2 := &fakeWsConn{closed: false}
		_, err = CM.NewConnection("test", ws2, h)
		So(err, ShouldBeNil)
		So(CM.ConnectionCount(), ShouldEqual, 1)

		conn1.Close()
		So(CM.ConnectionCount(), ShouldEqual, 1)
		So(ws2.closed, ShouldBeFalse)

		CM.Close()
		CM.Wait()
	})

	Convey("successfully closes all created connections.", t, func() {
		InitializeCM()

		concurrency := 100
		wss := make([]*fakeWsConn, 0)
		for i := 0; i < concurrency; i++ {
			wss = append(wss, &fakeWsConn{
				closed: false,
			})
		}

		for i := 0; i < concurrency; i++ {
			go func(iter int) {
				CM.NewConnection("test"+strconv.Itoa(iter), wss[iter], h)
			}(i)
		}

		CM.Close()
		CM.Wait()

		So(CM.ConnectionCount(), ShouldEqual, 0)

		allClosed := true
		for _, ws := range wss {
			if ws.closed == false {
				allClosed = false
			}
		}
		So(allClosed, ShouldBeTrue)
	})
}
