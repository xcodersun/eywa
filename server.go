package main

import (
	"fmt"
	// "github.com/spf13/viper"
	// "os"
	"time"
	// "path"
	"net/http"

	// . "github.com/vivowares/octopus.single/util"
	"github.com/gorilla/websocket"
)

func main() {
	// configure()
	// fmt.Println(viper.GetDuration("connections.expiry"))
	http.HandleFunc("/", handler)
	http.ListenAndServe(":8080", nil)
}

// func configure() {
// 	viper.SetConfigName("octopus.single")
// 	viper.AddConfigPath("/etc/octopus.single/")
// 	pwd, err := os.Getwd()
// 	PanicIfErr(err)
// 	viper.AddConfigPath(path.Join(pwd, "configs"))
// 	err = viper.ReadInConfig()
// 	if err != nil {
// 		PanicIfErr(err)
// 	}
// }

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func handler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	conn.SetPingHandler(func(p string) error {
		fmt.Println("ping")
		return nil
	})
	conn.SetPongHandler(func(p string) error {
		fmt.Println("pong")
		return nil
	})
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	t, p, err := conn.ReadMessage()
	fmt.Println(t, p, err)
}
