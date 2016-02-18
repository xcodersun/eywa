// +build integration

package api_tests

import (
	"fmt"
	"github.com/vivowares/octopus/Godeps/_workspace/src/github.com/bitly/go-simplejson"
	"github.com/vivowares/octopus/Godeps/_workspace/src/github.com/satori/go.uuid"
	. "github.com/vivowares/octopus/Godeps/_workspace/src/github.com/smartystreets/goconvey/convey"
	"github.com/vivowares/octopus/Godeps/_workspace/src/github.com/verdverm/frisby"
	. "github.com/vivowares/octopus/models"
	. "github.com/vivowares/octopus/utils"
	"log"
	"net/http"
	"os"
	"testing"
	"time"
)

func HttpUploadPath(channelId, deviceId string) string {
	return fmt.Sprintf("%s/channels/%s/devices/%s/upload", DeviceServer, channelId, deviceId)
}

func GetRawIndexPath(channelId string) string {
	return fmt.Sprintf("%s/channels/%s/raw", ApiServer, channelId)
}

func TestHttpUpload(t *testing.T) {

	InitializeDB()
	DB.LogMode(true)
	DB.SetLogger(log.New(os.Stdout, "", log.LstdFlags))
	DB.DropTableIfExists(&Channel{})
	DB.AutoMigrate(&Channel{})

	InitializeIndexClient()

	Convey("successfully uploads the structed data and indexed into ES via http", t, func() {
		startTime := NanoToMilli(time.Now().UnixNano())

		reqBody := Channel{
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
			SetJson(reqBody).Send()

		var chId string
		f.ExpectStatus(http.StatusCreated).
			AfterJson(func(F *frisby.Frisby, js *simplejson.Json, err error) {
			chId = js.MustMap()["id"].(string)
		})

		deviceId := "abc"
		tag1 := uuid.NewV4().String()
		data := map[string]interface{}{
			"tag1":   tag1,
			"tag2":   "monday",
			"field1": 100,
		}
		f = frisby.Create("http upload").Post(HttpUploadPath(chId, deviceId)).
			SetHeader("AccessToken", "token1").SetJson(data).Send()
		f.ExpectStatus(http.StatusOK)

		IndexClient.Refresh().Do()
		time.Sleep(3 * time.Second)

		f = frisby.Create("get raw index").Get(GetRawIndexPath(chId)).
			SetHeader("AuthToken", authStr()).
			SetParam("time_range", fmt.Sprintf("%d:", startTime)).
			SetParam("nop", "false").Send()

		f.ExpectStatus(http.StatusOK).
			AfterContent(func(F *frisby.Frisby, content []byte, err error) {
			js, _ := simplejson.NewJson(content)
			So(js.MustMap()["tag1"].(string), ShouldEqual, tag1)
		})

	})

	frisby.Global.PrintReport()
}
