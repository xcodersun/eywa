// +build integration

package api_tests

import (
	"encoding/json"
	"fmt"
	"github.com/vivowares/octopus/Godeps/_workspace/src/github.com/bitly/go-simplejson"
	. "github.com/vivowares/octopus/Godeps/_workspace/src/github.com/smartystreets/goconvey/convey"
	"github.com/vivowares/octopus/Godeps/_workspace/src/github.com/verdverm/frisby"
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

// import "github.com/kr/pretty"

var ApiServer string
var ConfigFile string

type ChannelResp struct {
	Id           string            `json:"id"`
	Name         string            `json:"name"`
	Description  string            `json:"description"`
	Tags         []string          `json:"tags"`
	Fields       map[string]string `json:"fields"`
	AccessTokens []string          `json:"access_tokens"`
}

func init() {
	pwd, err := os.Getwd()
	PanicIfErr(err)
	ConfigFile = path.Join(path.Dir(pwd), "configs", "octopus_test.yml")
	PanicIfErr(InitializeConfig(ConfigFile))

	ApiServer = "http://" + Config.Service.Host + ":" + strconv.Itoa(Config.Service.HttpPort)
}

func ListChannelPath() string {
	return fmt.Sprintf("%s/%s", ApiServer, "channels")
}

func GetChannelPath(base64Id string) string {
	return fmt.Sprintf("%s/%s/%s", ApiServer, "channels", base64Id)
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

		var chId string
		f.ExpectStatus(http.StatusOK).
			AfterJson(func(F *frisby.Frisby, js *simplejson.Json, err error) {
			So(len(js.MustArray()), ShouldEqual, 1)
			ch := js.MustArray()[0].(map[string]interface{})
			chId, _ = ch["id"].(string)
		})

		expResp := &ChannelResp{
			Id:           chId,
			Name:         reqBody.Name,
			Description:  reqBody.Description,
			Tags:         reqBody.Tags,
			Fields:       reqBody.Fields,
			AccessTokens: reqBody.AccessTokens,
		}

		f = frisby.Create("get channel").Get(GetChannelPath(chId)).Send()
		f.ExpectStatus(http.StatusOK).AfterContent(func(F *frisby.Frisby, resp []byte, err error) {
			ch := &ChannelResp{}
			json.Unmarshal(resp, ch)
			So(reflect.DeepEqual(ch, expResp), ShouldBeTrue)
		})

		f = frisby.Create("update channel").Put(GetChannelPath(chId))
		f.SetJson(map[string]string{"name": "updated name"}).Send()

		f.ExpectStatus(http.StatusOK)

		f = frisby.Create("get channel").Get(GetChannelPath(chId)).Send()
		expResp.Name = "updated name"
		f.ExpectStatus(http.StatusOK).AfterContent(func(F *frisby.Frisby, resp []byte, err error) {
			ch := &ChannelResp{}
			json.Unmarshal(resp, ch)
			So(reflect.DeepEqual(ch, expResp), ShouldBeTrue)
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
