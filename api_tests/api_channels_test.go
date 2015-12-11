package api_tests

import (
	"encoding/json"
	"fmt"
	"github.com/bitly/go-simplejson"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/verdverm/frisby"
	. "github.com/vivowares/octopus/configs"
	. "github.com/vivowares/octopus/models"
	. "github.com/vivowares/octopus/utils"
	"net/http"
	"os"
	"path"
	"reflect"
	"strconv"
	"testing"
)

var ApiServer string

func init() {
	pwd, err := os.Getwd()
	PanicIfErr(err)
	PanicIfErr(
		InitializeConfig(path.Join(path.Dir(pwd), "configs", "octopus_test.yml")),
	)

	ApiServer = "http://" + Config.Service.Host + ":" + strconv.Itoa(Config.Service.HttpPort)
}

func ListChannelPath() string {
	return fmt.Sprintf("%s/%s", ApiServer, "channels")
}

func GetChannelPath(id int64) string {
	return fmt.Sprintf("%s/%s/%d", ApiServer, "channels", id)
}

func TestApiChannels(t *testing.T) {
	InitializeDB()
	DB.DropTableIfExists(&Channel{})
	DB.AutoMigrate(&Channel{})

	frisby.Global.SetHeader("Content-Type", "application/json").SetHeader("Accept", "application/json")

	Convey("successfully creates/gets/lists/updates channel", t, func() {
		f := frisby.Create("list channels").Get(ListChannelPath()).Send()

		f.ExpectStatus(http.StatusOK).
			AfterJson(func(F *frisby.Frisby, js *simplejson.Json, err error) {
			So(len(js.MustArray()), ShouldEqual, 0)
		})

		reqBody := Channel{
			Name:         "test",
			Description:  "desc",
			Tags:         []string{"tag1", "tag2"},
			Fields:       map[string]string{"field1": "int"},
			AccessTokens: []string{"token1"},
		}
		f = frisby.Create("create channel").Post(ListChannelPath())
		f.SetJson(reqBody).Send()

		f.ExpectStatus(http.StatusCreated)

		f = frisby.Create("list channels").Get(ListChannelPath()).Send()

		var chId int64
		f.ExpectStatus(http.StatusOK).
			AfterJson(func(F *frisby.Frisby, js *simplejson.Json, err error) {
			So(len(js.MustArray()), ShouldEqual, 1)
			ch := js.MustArray()[0].(map[string]interface{})
			chId, _ = ch["id"].(json.Number).Int64()
		})

		f = frisby.Create("get channel").Get(GetChannelPath(chId)).Send()
		reqBody.Id = 1
		f.ExpectStatus(http.StatusOK).AfterContent(func(F *frisby.Frisby, resp []byte, err error) {
			ch := Channel{}
			json.Unmarshal(resp, &ch)
			So(reflect.DeepEqual(ch, reqBody), ShouldBeTrue)
		})

		f = frisby.Create("update channel").Put(GetChannelPath(chId))
		f.SetJson(map[string]string{"name": "updated name"}).Send()

		f.ExpectStatus(http.StatusOK)

		f = frisby.Create("get channel").Get(GetChannelPath(chId)).Send()
		reqBody.Name = "updated name"
		f.ExpectStatus(http.StatusOK).AfterContent(func(F *frisby.Frisby, resp []byte, err error) {
			ch := Channel{}
			json.Unmarshal(resp, &ch)
			So(reflect.DeepEqual(ch, reqBody), ShouldBeTrue)
		})

		f = frisby.Create("delete channel").Delete(GetChannelPath(chId)).Send()
		f.ExpectStatus(http.StatusOK)

		f = frisby.Create("list channels").Get(ListChannelPath()).Send()

		f.ExpectStatus(http.StatusOK).
			AfterJson(func(F *frisby.Frisby, js *simplejson.Json, err error) {
			So(len(js.MustArray()), ShouldEqual, 0)
		})
	})

	frisby.Global.PrintReport()
}
