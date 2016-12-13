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
			SetHeader("Api-Key", Config().Security.ApiKey).SetJson(map[string]string{"test": message}).Send()
		f.ExpectStatus(http.StatusOK)

		wg.Wait()
		So(rcvErr, ShouldBeNil)
		So(rcvMsgType, ShouldEqual, websocket.BinaryMessage)
		strs := strings.Split(string(rcvData), "|")
		So(strs[len(strs)-1], ShouldEqual, fmt.Sprintf("{\"test\":\"%s\"}", message))
		So(strs[0], ShouldEqual, strconv.Itoa(int(TypeSendMessage)))

		cli.Close()
	})

	Convey("successfully request data from device", t, func() {
		deviceId := "abc"
		cli := CreateWsConnection(chId, deviceId, ch)

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

		f := frisby.Create("send message to device").Post(ApiRequestToDevicePath(chId, deviceId)).
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

		cli.Close()
	})

	Convey("successfully uploads the structed data and indexed into ES via http, also long polling for downloading data", t, func() {
		reqBody := Channel{
			Name:         "test http polling",
			Description:  "desc",
			Tags:         []string{"tag1", "tag2"},
			Fields:       map[string]string{"field1": "int"},
			AccessTokens: []string{"token1"},
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

	DeleteTestChannel(chId)

	frisby.Global.PrintReport()
}
