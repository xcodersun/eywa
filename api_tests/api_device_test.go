// +build integration

package api_tests

import (
	"fmt"
	"github.com/bitly/go-simplejson"
	"github.com/gorilla/websocket"
	"github.com/satori/go.uuid"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/verdverm/frisby"
	"gopkg.in/olivere/elastic.v3"
	. "github.com/eywa/configs"
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

func ApiSendToDevicePath(chId, deviceId string) string {
	return fmt.Sprintf("%s/api/channels/%s/devices/%s/send", ApiServer, chId, deviceId)
}

func ApiRequestToDevicePath(chId, deviceId string) string {
	return fmt.Sprintf("%s/api/channels/%s/devices/%s/request", ApiServer, chId, deviceId)
}

func TestApiToDevice(t *testing.T) {
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

	Convey("Send data to device from api", t, func() {
		message := "this is a test message"
		var rcvData []byte
		var rcvMsgType int
		var rcvErr error
		var wg sync.WaitGroup

		// Increment a WaitGroup counter which will be decremented once client
		// receives a message from server.
		wg.Add(1)
		// Start an async thread to wait for message from server.
		go func() {
			cli.SetReadDeadline(time.Now().Add(2 * time.Second))
			rcvMsgType, rcvData, rcvErr = cli.ReadMessage()
			// A message is received so release the counter.
			wg.Done()
		}()

		// send {"test": "this is a test message"} to
		// http://localhost:9090/api/channels/<chId>/devices/<deviceId>/send
		f := frisby.Create("send message to device").Post(ApiSendToDevicePath(chId, websocketDeviceId)).
			SetHeader("Api-Key", Config().Security.ApiKey).SetJson(map[string]string{"test": message}).Send()
		f.ExpectStatus(http.StatusOK)

		// All the error checks is blocked until the async thread decrement the
		// WaitGroup counter.
		wg.Wait()
		So(rcvErr, ShouldBeNil)
		So(rcvMsgType, ShouldEqual, websocket.BinaryMessage)
		strs := strings.Split(string(rcvData), "|")
		So(strs[len(strs)-1], ShouldEqual, fmt.Sprintf("{\"test\":\"%s\"}", message))
		So(strs[0], ShouldEqual, strconv.Itoa(int(TypeSendMessage)))
	})

	Convey("Request data from device", t, func() {
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

		f := frisby.Create("send message to device").Post(ApiRequestToDevicePath(chId, websocketDeviceId)).
			SetHeader("Api-Key", Config().Security.ApiKey).SetJson(map[string]string{"test": reqMsg}).Send()
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

	Convey("successfully uploads the structed data and indexed into ES via http, also long polling for downloading data", t, func() {
		reqBody := Channel{
			Name:            "test http polling",
			Description:     "desc",
			Tags:            []string{"tag1", "tag2"},
			Fields:          map[string]string{"field1": "int"},
			AccessTokens:    []string{"token1"},
			ConnectionLimit: 5,
			MessageRate:     1000,
		}
		f := frisby.Create("create channel").Post(ListChannelPath()).
			SetHeader("Content-Type", "application/json").
			SetHeader("Accept", "application/json").
			SetHeader("Authentication", authStr()).
			SetJson(reqBody).Send()

		var chId string
		f.ExpectStatus(http.StatusCreated).
			AfterJson(func(F *frisby.Frisby, js *simplejson.Json, err error) {
			chId = js.MustMap()["id"].(string)
		})

		deviceId := uuid.NewV4().String()
		tag1 := uuid.NewV4().String()
		data := map[string]interface{}{
			"tag1":   tag1,
			"tag2":   "monday",
			"field1": 100,
		}

		var response string
		go func() {
			f = frisby.Create("http long polling").Get(HttpPollingPath(chId, deviceId)).
				SetHeader("AccessToken", "token1").SetJson(data).Send()
			f.ExpectStatus(http.StatusOK).AfterContent(func(F *frisby.Frisby, content []byte, err error) {
				response = string(content)
			})
		}()
		time.Sleep(1 * time.Second)
		f = frisby.Create("send message to device").Post(ApiSendToDevicePath(chId, deviceId)).
			SetHeader("Api-Key", Config().Security.ApiKey).SetJson(map[string]string{"test": "message"}).Send()
		f.ExpectStatus(http.StatusOK)

		IndexClient.Refresh().Do()
		time.Sleep(3 * time.Second)
		So(response, ShouldEqual, `{"test":"message"}`)

		searchRes, err := IndexClient.Search().Index("_all").Type(IndexTypeMessages).Query(elastic.NewTermQuery("tag1", tag1)).Do()
		So(err, ShouldBeNil)
		So(searchRes.TotalHits(), ShouldEqual, 1)

		searchRes, err = IndexClient.Search().Index("_all").Type(IndexTypeActivities).Query(elastic.NewTermQuery("device_id", deviceId)).Do()
		So(err, ShouldBeNil)
		So(searchRes.TotalHits(), ShouldEqual, 2)
	})
	// Delete the channel
	DeleteTestChannel(chId)

	frisby.Global.PrintReport()
}
