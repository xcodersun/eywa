// +build integration

package api_tests

import (
	"fmt"
	"github.com/satori/go.uuid"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/verdverm/frisby"
	"gopkg.in/olivere/elastic.v3"
	. "github.com/eywa/models"
	"log"
	"net/http"
	"os"
	"testing"
	"time"
)

func HttpUploadPath(channelId, deviceId string) string {
	return fmt.Sprintf("%s/channels/%s/devices/%s/upload", DeviceServer, channelId, deviceId)
}

func HttpPollingPath(channelId, deviceId string) string {
	return fmt.Sprintf("%s/channels/%s/devices/%s/poll", DeviceServer, channelId, deviceId)
}

func GetRawIndexPath(channelId string) string {
	return fmt.Sprintf("%s/channels/%s/raw", ApiServer, channelId)
}

func TestHttpConnection(t *testing.T) {

	InitializeDB()
	DB.LogMode(true)
	DB.SetLogger(log.New(os.Stdout, "", log.LstdFlags))
	DB.DropTableIfExists(&Channel{})
	DB.AutoMigrate(&Channel{})

	InitializeIndexClient()

	chId, _ := CreateTestChannel()

	Convey("successfully uploads the structed data and indexed into ES via http", t, func() {

		deviceId := "abc"
		tag1 := uuid.NewV4().String()
		data := map[string]interface{}{
			"tag1":   tag1,
			"tag2":   "monday",
			"field1": 100,
		}
		f := frisby.Create("http upload").Post(HttpUploadPath(chId, deviceId)).
			SetHeader("AccessToken", "token1").SetJson(data).Send()
		f.ExpectStatus(http.StatusOK)

		IndexClient.Refresh().Do()
		time.Sleep(3 * time.Second)

		searchRes, err := IndexClient.Search().Index("_all").Query(elastic.NewTermQuery("tag1", tag1)).Do()
		So(err, ShouldBeNil)
		So(searchRes.TotalHits(), ShouldEqual, 1)
	})

	Convey("data won't be indexed if indices.disable is enabled", t, func() {
		f := frisby.Create("disable index").SetHeader("Authentication", authStr()).
			Put(ConfigsPath()).SetJson(map[string]interface{}{"indices": map[string]interface{}{"disable": true}}).Send()
		f.ExpectStatus(http.StatusOK)

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

		searchRes, err := IndexClient.Search().Index("_all").Query(elastic.NewTermQuery("tag1", tag1)).Do()
		So(err, ShouldBeNil)
		So(searchRes.TotalHits(), ShouldEqual, 0)

		f = frisby.Create("enable index").SetHeader("Authentication", authStr()).
			Put(ConfigsPath()).SetJson(map[string]interface{}{"indices": map[string]interface{}{"disable": false}}).Send()
		f.ExpectStatus(http.StatusOK)
	})

	DeleteTestChannel(chId)

	frisby.Global.PrintReport()
}
