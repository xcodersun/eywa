package connections

import (
	. "github.com/vivowares/octopus/Godeps/_workspace/src/github.com/smartystreets/goconvey/convey"
	. "github.com/vivowares/octopus/configs"
	"math/rand"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestRaceConditions(t *testing.T) {

	SetConfig(&Conf{
		WebSocketConnections: &WsConnectionConf{
			Registry:         "memory",
			NShards:          4,
			InitShardSize:    8,
			RequestQueueSize: 8,
			Expiry:           300 * time.Second,
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

	Convey("burst various sends for race condition test, with wg", t, func() {
		InitializeWSCM()
		defer CloseWSCM()
		ws := &fakeWsConn{randomErr: false}
		conn, _ := NewWebSocketConnection("test", ws, h, meta)

		concurrency := 1000
		var wg sync.WaitGroup
		wg.Add(concurrency)
		errs := make([]error, concurrency)
		for i := 0; i < concurrency; i++ {
			go func(index int) {
				var msg []byte
				var err error
				switch rand.Intn(3) {
				case 0:
					msg = []byte("async" + strconv.Itoa(index))
					err = conn.SendAsyncRequest(msg)
				case 1:
					msg = []byte("resp" + strconv.Itoa(index))
					err = conn.SendResponse(msg)
				case 2:
					msg = []byte("sync" + strconv.Itoa(index))
					_, err = conn.SendSyncRequest(msg)
				}
				errs[index] = err
				wg.Done()
			}(i)
		}

		wg.Wait()
		conn.Close()
		conn.Wait()
		So(WebSocketCount(), ShouldEqual, 0)

		So(ws.closed, ShouldBeTrue)
		So(conn.msgChans.len(), ShouldEqual, 0) //?
		hasClosedConnErr := false
		for _, err := range errs {
			if err != nil && strings.Contains(err.Error(), "connection is closed") {
				hasClosedConnErr = true
			}
		}
		So(hasClosedConnErr, ShouldBeFalse)
	})

	Convey("burst various sends for race condition test, without wg", t, func() {
		InitializeWSCM()
		ws := &fakeWsConn{randomErr: false}
		conn, _ := NewWebSocketConnection("test", ws, h, meta)

		concurrency := 1000
		errs := make([]error, concurrency)
		for i := 0; i < concurrency; i++ {
			go func(index int) {
				var msg []byte
				var err error
				switch rand.Intn(3) {
				case 0:
					msg = []byte("async" + strconv.Itoa(index))
					err = conn.SendAsyncRequest(msg)
				case 1:
					msg = []byte("resp" + strconv.Itoa(index))
					err = conn.SendResponse(msg)
				case 2:
					msg = []byte("sync" + strconv.Itoa(index))
					_, err = conn.SendSyncRequest(msg)
				}
				errs[index] = err
			}(i)
		}

		CloseWSCM()
		So(WebSocketCount(), ShouldEqual, 0)
		So(ws.closed, ShouldBeTrue)
	})

	Convey("successfully closes all created connections.", t, func() {
		InitializeWSCM()

		concurrency := 100
		wss := make([]*fakeWsConn, concurrency)
		for i := 0; i < concurrency; i++ {
			wss[i] = &fakeWsConn{}
		}
		var wg sync.WaitGroup
		wg.Add(concurrency)
		for i := 0; i < concurrency; i++ {
			go func(iter int) {
				NewWebSocketConnection("test"+strconv.Itoa(iter), wss[iter], h, meta)
				wg.Done()
			}(i)
		}
		CloseWSCM()
		wg.Wait()

		So(WebSocketCount(), ShouldEqual, 0)

		allClosed := true
		for _, ws := range wss {
			if ws.closed == false {
				allClosed = false
			}
		}
		So(allClosed, ShouldBeTrue)
	})

	Convey("real life race conditions, close all underlying ws conn.", t, func() {
		concurrency := 1000
		InitializeWSCM()
		wss := make([]*fakeWsConn, concurrency)
		for i := 0; i < concurrency; i++ {
			wss[i] = &fakeWsConn{randomErr: rand.Intn(4) == 0}
		}
		conns := make([]*WebSocketConnection, concurrency)
		errs := make([]error, concurrency)
		for i := 0; i < concurrency; i++ {
			go func(iter int) {
				time.Sleep(time.Duration(rand.Intn(500)) * time.Millisecond)
				conn, err := NewWebSocketConnection("test"+strconv.Itoa(iter), wss[iter], h, meta)
				conns[iter] = conn
				errs[iter] = err
				switch rand.Intn(3) {
				case 0:
					conn.SendAsyncRequest([]byte("async" + strconv.Itoa(iter)))
				case 1:
					conn.SendResponse([]byte("resp" + strconv.Itoa(iter)))
				case 2:
					conn.SendSyncRequest([]byte("sync" + strconv.Itoa(iter)))
				}
			}(i)
		}

		time.Sleep(time.Duration(200+rand.Intn(500)) * time.Millisecond)
		CloseWSCM()
		So(WebSocketCount(), ShouldEqual, 0)

		time.Sleep(time.Duration(1+rand.Intn(3)) * time.Second)
		allClosed := true
		for i, ws := range wss {
			if errs[i] == nil && ws.closed == false {
				allClosed = false
			}
		}
		So(allClosed, ShouldBeTrue)
	})
}
