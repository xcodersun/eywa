package connections

import (
	"bytes"
	"fmt"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/spf13/viper"
	"strconv"
	"testing"
	"time"
)

func TestConnectionManager(t *testing.T) {
	viper.SetConfigType("yaml")

	var yaml = []byte(`
    connections:
      store: memory
      expiry: &expiry 1s
      timeouts:
        write: 2s
        read: *expiry
        response: 8s
      buffer_sizes:
        read: 1024
        write: 1024
  `)

	viper.ReadConfig(bytes.NewBuffer(yaml))

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

		conn1.close()
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
		fmt.Println(len(wss))
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
