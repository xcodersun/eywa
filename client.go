package main

import (
	"fmt"
	"github.com/gorilla/websocket"
	"net"
	"net/http"
	"net/url"
	// "os"
	"time"
)

func main() {
	u, err := url.Parse("ws://localhost:8080/")
	if err != nil {
		return
	}

	rawConn, err := net.Dial("tcp", u.Host)
	if err != nil {
		return
	}

	wsHeaders := http.Header{
		"Origin": {"ws://localhost:8080"},
		// your milage may differ
		"Sec-WebSocket-Extensions": {"permessage-deflate; client_max_window_bits, x-webkit-deflate-frame"},
	}

	conn, resp, err := websocket.NewClient(rawConn, u, wsHeaders, 1024, 1024)
	if err != nil {
		fmt.Errorf("websocket.NewClient Error: %s\nResp:%+v", err, resp)
	}

	conn.WriteMessage(websocket.CloseMessage, []byte("hello"))
	time.Sleep(3 * time.Second)
	fmt.Println(conn.ReadMessage())
	time.Sleep(20 * time.Second)
	fmt.Println("done")
}
