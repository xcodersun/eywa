// +build integration

package api_tests

import (
	"encoding/json"
	"fmt"
	"github.com/vivowares/octopus/Godeps/_workspace/src/github.com/bitly/go-simplejson"
	"github.com/vivowares/octopus/Godeps/_workspace/src/github.com/gorilla/websocket"
	"github.com/vivowares/octopus/Godeps/_workspace/src/github.com/satori/go.uuid"
	. "github.com/vivowares/octopus/Godeps/_workspace/src/github.com/smartystreets/goconvey/convey"
	"github.com/vivowares/octopus/Godeps/_workspace/src/github.com/verdverm/frisby"
	. "github.com/vivowares/octopus/configs"
	. "github.com/vivowares/octopus/models"
	. "github.com/vivowares/octopus/utils"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"testing"
	"time"
)

func CreateTestChannel() (string, *Channel) {
	ch := &Channel{
		Name:            "test http upload",
		Description:     "desc",
		Tags:            []string{"tag1", "tag2"},
		Fields:          map[string]string{"field1": "int"},
		MessageHandlers: []string{"indexer"},
		AccessTokens:    []string{"token1"},
	}
	f := frisby.Create("create channel").Post(ListChannelPath()).
		SetHeader("Content-Type", "application/json").
		SetHeader("Accept", "application/json").
		SetHeader("AuthToken", authStr()).
		SetJson(ch).Send()

	var chId string
	f.ExpectStatus(http.StatusCreated).
		AfterJson(func(F *frisby.Frisby, js *simplejson.Json, err error) {
		chId = js.MustMap()["id"].(string)
	})

	return chId, ch
}

func CreateWsConnection(chId, deviceId string, ch *Channel) *websocket.Conn {
	u := url.URL{
		Scheme: "ws",
		Host: fmt.Sprintf("%s:%d",
			Config().Service.Host, Config().Service.DevicePort),
		Path: fmt.Sprintf("/ws/channels/%s/devices/%s", chId, deviceId),
	}
	h := map[string][]string{"AccessToken": ch.AccessTokens}

	cli, _, err := websocket.DefaultDialer.Dial(u.String(), h)
	So(err, ShouldBeNil)
	return cli
}

func ConnectionCountPath() string {
	return fmt.Sprintf("%s/%s", ApiServer, "ws/connections/_count")
}

func CheckConnectionCount() int64 {
	f := frisby.Create("check connection count").Get(ConnectionCountPath()).
		SetHeader("Content-Type", "application/json").
		SetHeader("Accept", "application/json").
		SetHeader("AuthToken", authStr()).
		Send()

	var count int64
	f.ExpectStatus(http.StatusOK).
		AfterJson(func(F *frisby.Frisby, js *simplejson.Json, err error) {
		count, _ = js.MustMap()["count"].(json.Number).Int64()
	})
	return count
}

func TestWsConnection(t *testing.T) {

	InitializeDB()
	DB.LogMode(true)
	DB.SetLogger(log.New(os.Stdout, "", log.LstdFlags))
	DB.DropTableIfExists(&Channel{})
	DB.AutoMigrate(&Channel{})

	InitializeIndexClient()

	chId, ch := CreateTestChannel()

	Convey("successfully ping the server and get the timestamp", t, func() {
		cli := CreateWsConnection(chId, "abc", ch)
		So(CheckConnectionCount(), ShouldEqual, 1)
		//test ping
		cli.SetPongHandler(func(data string) error {
			_, err := strconv.ParseInt(data, 10, 64)
			So(err, ShouldBeNil)
			return nil
		})
		cli.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(1*time.Second))
		cli.SetReadDeadline(time.Now().Add(1 * time.Second))
		cli.ReadMessage()

		cli.Close()
		So(CheckConnectionCount(), ShouldEqual, 0)
	})

	Convey("successfully uploads structured data and get it indexed", t, func() {
		cli := CreateWsConnection(chId, "abc", ch)
		So(CheckConnectionCount(), ShouldEqual, 1)

		startTime := NanoToMilli(time.Now().UnixNano())
		tag1 := uuid.NewV4().String()
		data := fmt.Sprintf("1|123|tag1=%s&field1=100", tag1)
		cli.SetWriteDeadline(time.Now().Add(1 * time.Second))
		cli.WriteMessage(websocket.BinaryMessage, []byte(data))

		IndexClient.Refresh().Do()
		time.Sleep(3 * time.Second)

		f := frisby.Create("get raw index").Get(GetRawIndexPath(chId)).
			SetHeader("AuthToken", authStr()).
			SetParam("time_range", fmt.Sprintf("%d:", startTime)).
			SetParam("nop", "false").Send()

		f.ExpectStatus(http.StatusOK).
			AfterContent(func(F *frisby.Frisby, content []byte, err error) {
			js, _ := simplejson.NewJson(content)
			So(js.MustMap()["tag1"].(string), ShouldEqual, tag1)
		})

		cli.Close()
		So(CheckConnectionCount(), ShouldEqual, 0)
	})

	frisby.Global.PrintReport()
}
