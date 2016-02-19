// +build integration

package api_tests

import (
	"fmt"
	// "github.com/vivowares/octopus/Godeps/_workspace/src/github.com/bitly/go-simplejson"
	// "github.com/vivowares/octopus/Godeps/_workspace/src/github.com/gorilla/websocket"
	. "github.com/vivowares/octopus/Godeps/_workspace/src/github.com/smartystreets/goconvey/convey"
	"github.com/vivowares/octopus/Godeps/_workspace/src/github.com/verdverm/frisby"
	// . "github.com/vivowares/octopus/configs"
	. "github.com/vivowares/octopus/connections"
	. "github.com/vivowares/octopus/models"
	// . "github.com/vivowares/octopus/utils"
	"log"
	// "net/http"
	// "net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"
)

import "github.com/kr/pretty"

func ApiSendToDevicePath(chId, deviceId string) string {
	return fmt.Sprintf("%s/channels/%s/devices/%s/send", ApiServer, chId, deviceId)
}

func TestApiToDevice(t *testing.T) {

	InitializeDB()
	DB.LogMode(true)
	DB.SetLogger(log.New(os.Stdout, "", log.LstdFlags))
	DB.DropTableIfExists(&Channel{})
	DB.AutoMigrate(&Channel{})

	InitializeIndexClient()

	chId, ch := CreateTestChannel()
	pretty.Println("((((((((((((((((((((((((((")
	pretty.Println(chId)
	pretty.Println(ch)

	Convey("successfully send data to device from api", t, func() {
		deviceId := "abc"
		cli := CreateWsConnection(chId, deviceId, ch)

		message := "this is a test message"
		var rcvData []byte
		var rcvMsgType int
		var rcvErr error
		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			cli.SetReadDeadline(time.Now().Add(2 * time.Second))
			rcvMsgType, rcvData, rcvErr = cli.ReadMessage()
			wg.Done()
		}()

		f := frisby.Create("send message to device").Post(ApiSendToDevicePath(chId, deviceId)).
			SetHeader("AuthToken", authStr()).SetJson(map[string]string{"test": message}).Send()
		f.
			AfterContent(func(F *frisby.Frisby, content []byte, err error) {
			pretty.Println(string(content))
		})

		pretty.Println(")))))))))))))))))))))))")
		pretty.Println(string(rcvData))
		pretty.Println(rcvErr)
		// So(rcvErr, ShouldBeNil)
		// So(rcvMsgType, ShouldEqual, websocket.BinaryMessage)
		strs := strings.Split(string(rcvData), "|")
		So(strs[len(strs)-1], ShouldEqual, fmt.Sprintf("{\"test\":\"%s\"}", message))
		So(strs[0], ShouldEqual, strconv.Itoa(SendMessage))

		wg.Wait()
		cli.Close()
	})

	// Convey("successfully request data to device from api", t, func() {
	// 	deviceId := "abc"
	// 	cli := CreateWsConnection(chId, deviceId, ch)
	// 	So(CheckConnectionCount(), ShouldEqual, 1)

	// 	message := "this is a test message"
	// 	var rcvData []byte
	// 	var rcvMsgType int
	// 	var rcvErr error
	// 	var wg sync.WaitGroup
	// 	wg.Add(1)
	// 	go func() {
	// 		cli.SetReadDeadline(time.Now().Add(2 * time.Second))
	// 		rcvMsgType, rcvData, rcvErr = cli.ReadMessage()
	// 		wg.Done()
	// 	}()

	// 	frisby.Create("send message to device").Post(ApiSendToDevicePath(chId, deviceId)).
	// 		SetHeader("AuthToken", authStr()).SetJson(map[string]string{"test": message}).Send()

	// 	So(rcvErr, ShouldBeNil)
	// 	So(rcvMsgType, ShouldEqual, websocket.BinaryMessage)
	// 	strs := strings.Split(string(rcvData), "|")
	// 	So(strs[len(strs)-1], ShouldEqual, fmt.Sprintf("{\"test\":\"%s\"}", message))
	// 	So(strs[len(strs)-1], ShouldEqual, strconv.Itoa(SendMessage))

	// 	wg.Wait()
	// 	cli.Close()
	// 	So(CheckConnectionCount(), ShouldEqual, 0)
	// })

	frisby.Global.PrintReport()
}
