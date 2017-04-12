// +build integration

package api_tests

import (
	"encoding/json"
	"fmt"
	"github.com/bitly/go-simplejson"
	"github.com/gorilla/websocket"
	"github.com/satori/go.uuid"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/verdverm/frisby"
	"gopkg.in/olivere/elastic.v3"
	. "github.com/eywa/configs"
	. "github.com/eywa/models"
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
		Name:            "test channel",
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
		SetJson(ch).Send()

	var chId string
	f.ExpectStatus(http.StatusCreated).
		AfterJson(func(F *frisby.Frisby, js *simplejson.Json, err error) {
		chId = js.MustMap()["id"].(string)
	})

	return chId, ch
}

func DeleteTestChannel(chId string) {
	f := frisby.Create("delete channel").Delete(GetChannelPath(chId)).SetHeader("Authentication", authStr()).Send()
	f.ExpectStatus(http.StatusOK)
}

func CreateWsConnection(chId, deviceId string, ch *Channel) (*websocket.Conn, error) {
	u := url.URL{
		Scheme: "ws",
		Host: fmt.Sprintf("%s:%d",
			Config().Service.Host, Config().Service.DevicePort),
		Path: fmt.Sprintf("/channels/%s/devices/%s/ws", chId, deviceId),
	}
	h := map[string][]string{"AccessToken": ch.AccessTokens}

	cli, _, err := websocket.DefaultDialer.Dial(u.String(), h)
	return cli, err
}

func ConnectionCountPath(chId string) string {
	return fmt.Sprintf("%s/admin/channels/%s/connections/count", ApiServer, chId)
}

func CheckConnectionCount(chId string) int64 {
	f := frisby.Create("check connection count").Get(ConnectionCountPath(chId)).
		SetHeader("Content-Type", "application/json").
		SetHeader("Accept", "application/json").
		SetHeader("Authentication", authStr()).
		Send()

	var count int64
	f.ExpectStatus(http.StatusOK).
		AfterJson(func(F *frisby.Frisby, js *simplejson.Json, err error) {
		count, _ = js.MustMap()[chId].(json.Number).Int64()
	})
	return count
}

func TestWsConnection(t *testing.T) {
	// Initialize database
	InitializeDB()
	DB.LogMode(true)
	DB.SetLogger(log.New(os.Stdout, "", log.LstdFlags))
	DB.DropTableIfExists(&Channel{})
	DB.AutoMigrate(&Channel{})
	// Initialize elasticsearch index client
	InitializeIndexClient()
	// Create a test channel
	chId, ch := CreateTestChannel()

	// Create websocket connection
	cli, cliErr := CreateWsConnection(chId, "abc", ch)
	// Wait for a few seconds for connection manager to register the new connection
	time.Sleep(2 * time.Second)

	Convey("Websocket connection is ready", t, func() {
		So(cliErr, ShouldBeNil)
		So(CheckConnectionCount(chId), ShouldEqual, 1)
	})

	Convey("successfully ping the server and get the timestamp", t, func() {
		cli.SetPongHandler(func(data string) error {
			_, err := strconv.ParseInt(data, 10, 64)
			So(err, ShouldBeNil)
			return nil
		})
		cli.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(1*time.Second))
		cli.SetReadDeadline(time.Now().Add(1 * time.Second))
		// Wait for ping from server
		cli.ReadMessage()
	})

	Convey("successfully uploads structured data and get it indexed", t, func() {
		tag1 := uuid.NewV4().String()
		data := fmt.Sprintf("1|123|tag1=%s&field1=100", tag1)
		cli.SetWriteDeadline(time.Now().Add(1 * time.Second))
		cli.WriteMessage(websocket.BinaryMessage, []byte(data))

		IndexClient.Refresh().Do()
		time.Sleep(3 * time.Second)
		searchRes, err := IndexClient.Search().Index("_all").Query(elastic.NewTermQuery("tag1", tag1)).Do()
		So(err, ShouldBeNil)
		So(searchRes.TotalHits(), ShouldEqual, 1)
	})

	Convey("Close the websocket connection", t, func() {
		cli.Close()
		time.Sleep(2 * time.Second)
		So(CheckConnectionCount(chId), ShouldEqual, 0)
	})
	// Delete the channel
	DeleteTestChannel(chId)
	frisby.Global.PrintReport()
}
