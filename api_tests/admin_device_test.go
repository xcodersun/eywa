// +build integration

package api_tests

import (
	"fmt"
	"github.com/gorilla/websocket"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/verdverm/frisby"
	. "github.com/eywa/connections"
	. "github.com/eywa/models"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"
)

func AdminSendToDevicePath(chId, deviceId string) string {
	return fmt.Sprintf("%s/admin/channels/%s/devices/%s/send", ApiServer, chId, deviceId)
}

func AdminRequestToDevicePath(chId, deviceId string) string {
	return fmt.Sprintf("%s/admin/channels/%s/devices/%s/request", ApiServer, chId, deviceId)
}

func TestAdminToDevice(t *testing.T) {

	InitializeDB()
	DB.LogMode(true)
	DB.SetLogger(log.New(os.Stdout, "", log.LstdFlags))
	DB.DropTableIfExists(&Channel{})
	DB.AutoMigrate(&Channel{})

	InitializeIndexClient()

	chId, ch := CreateTestChannel()

	websocketDeviceId := "abc"
	// Create websocket connection
	cli, cliErr := CreateWsConnection(chId, websocketDeviceId, ch)
	// Wait for a few seconds for connection manager to register the new connection
	time.Sleep(2 * time.Second)

	Convey("Websocket connection is ready", t, func() {
		So(cliErr, ShouldBeNil)
		So(CheckConnectionCount(chId), ShouldEqual, 1)
	})

	Convey("successfully send data to device from api", t, func() {
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

		f := frisby.Create("send message to device").Post(AdminSendToDevicePath(chId, websocketDeviceId)).
			SetHeader("Authentication", authStr()).SetJson(map[string]string{"test": message}).Send()
		f.ExpectStatus(http.StatusOK)

		wg.Wait()

		So(rcvErr, ShouldBeNil)
		So(rcvMsgType, ShouldEqual, websocket.BinaryMessage)
		strs := strings.Split(string(rcvData), "|")
		So(strs[len(strs)-1], ShouldEqual, fmt.Sprintf("{\"test\":\"%s\"}", message))
		So(strs[0], ShouldEqual, strconv.Itoa(int(TypeSendMessage)))
	})

	Convey("successfully request data from device", t, func() {
		reqMsg := "request message"
		respMsg := "response message"
		var rcvMessage string
		var rcvData []byte
		var rcvMsgType int
		var rcvErr error
		var sendMsgType MessageType
		var wg sync.WaitGroup

		wg.Add(1)
		go func() {
			cli.SetReadDeadline(time.Now().Add(2 * time.Second))
			rcvMsgType, rcvData, rcvErr = cli.ReadMessage()
			msg := NewWebsocketMessage(TypeRequestMessage, "", nil, rcvData)
			msg.Unmarshal()
			rcvMessage = string(msg.Payload())
			sendMsgType = msg.Type()
			msg = NewWebsocketMessage(TypeResponseMessage, msg.Id(), []byte(respMsg), nil)
			p, _ := msg.Marshal()
			cli.WriteMessage(websocket.BinaryMessage, p)
			wg.Done()
		}()

		f := frisby.Create("send message to device").Post(AdminRequestToDevicePath(chId, websocketDeviceId)).
			SetHeader("Authentication", authStr()).SetJson(map[string]string{"test": reqMsg}).Send()
		f.ExpectStatus(http.StatusOK).
			AfterContent(func(F *frisby.Frisby, content []byte, err error) {
			So(string(content), ShouldEqual, respMsg)
		})

		wg.Wait()

		So(rcvErr, ShouldBeNil)
		So(rcvMsgType, ShouldEqual, websocket.BinaryMessage)
		So(rcvMessage, ShouldEqual, `{"test":"request message"}`)
		So(sendMsgType, ShouldEqual, TypeRequestMessage)
	})

	Convey("Close the websocket connection", t, func() {
		cli.Close()
		time.Sleep(2 * time.Second)
		So(CheckConnectionCount(chId), ShouldEqual, 0)
	})

	DeleteTestChannel(chId)

	frisby.Global.PrintReport()
}
