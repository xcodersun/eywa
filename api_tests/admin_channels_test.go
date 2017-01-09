// +build integration

package api_tests

import (
	"encoding/json"
	"fmt"
	"github.com/bitly/go-simplejson"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/verdverm/frisby"
	. "github.com/eywa/configs"
	. "github.com/eywa/models"
	. "github.com/eywa/utils"
	"log"
	"net/http"
	"os"
	"path"
	"reflect"
	"strconv"
	"testing"
)

var ApiServer string
var DeviceServer string
var ConfigFile string

type ChannelResp struct {
	Id              string            `json:"id"`
	Name            string            `json:"name"`
	Description     string            `json:"description"`
	Tags            []string          `json:"tags"`
	Fields          map[string]string `json:"fields"`
	AccessTokens    []string          `json:"access_tokens"`
	ConnectionLimit int               `json:"connection_limit"`
	MessageRate     int               `json:"message_rate"`
}

func init() {
	pwd, err := os.Getwd()
	FatalIfErr(err)
	ConfigFile = path.Join(path.Dir(pwd), "configs", "eywa_test.yml")
	eywaHome := os.Getenv("EYWA_HOME")
	if len(eywaHome) == 0 {
		log.Fatalln("EYWA_HOME is not set")
	}
	params := map[string]string{"eywa_home": eywaHome}
	FatalIfErr(InitializeConfig(ConfigFile, params))

	ApiServer = "http://" + Config().Service.Host + ":" + strconv.Itoa(Config().Service.ApiPort)
	DeviceServer = "http://" + Config().Service.Host + ":" + strconv.Itoa(Config().Service.DevicePort)
}

func authStr() string {
	auth, err := NewAuthToken(
		Config().Security.Dashboard.Username,
		Config().Security.Dashboard.Password,
	)
	FatalIfErr(err)

	str, err := auth.Encrypt()
	FatalIfErr(err)
	return str
}

func ListChannelPath() string {
	return fmt.Sprintf("%s/%s", ApiServer, "admin/channels")
}

func GetChannelPath(hashId string) string {
	return fmt.Sprintf("%s/%s/%s", ApiServer, "admin/channels", hashId)
}

func ConnectionCountsPath() string {
	return fmt.Sprintf("%s/%s", ApiServer, "admin/connections/counts")
}

func TestAdminChannels(t *testing.T) {

	InitializeDB()
	DB.LogMode(true)
	DB.SetLogger(log.New(os.Stdout, "", log.LstdFlags))
	DB.DropTableIfExists(&Channel{})
	DB.AutoMigrate(&Channel{})

	frisby.Global.SetHeader("Content-Type", "application/json").
		SetHeader("Accept", "application/json").
		SetHeader("Authentication", authStr())

	Convey("successfully creates/gets/lists/updates channel", t, func() {
		f := frisby.Create("list channels").Get(ListChannelPath()).Send()

		f.ExpectStatus(http.StatusOK).
			AfterJson(func(F *frisby.Frisby, js *simplejson.Json, err error) {
			So(len(js.MustArray()), ShouldEqual, 0)
		})

		reqBody := Channel{
			Name:            "test",
			Description:     "desc",
			Tags:            []string{"tag1", "tag2"},
			Fields:          map[string]string{"field1": "int"},
			AccessTokens:    []string{"token1"},
			ConnectionLimit: 5,
			MessageRate:     1000,
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
			Id:              chId,
			Name:            reqBody.Name,
			Description:     reqBody.Description,
			Tags:            reqBody.Tags,
			Fields:          reqBody.Fields,
			AccessTokens:    reqBody.AccessTokens,
			ConnectionLimit: reqBody.ConnectionLimit,
			MessageRate:     reqBody.MessageRate,
		}

		f = frisby.Create("get channel").Get(GetChannelPath(chId)).Send()
		f.ExpectStatus(http.StatusOK).AfterContent(func(F *frisby.Frisby, resp []byte, err error) {
			ch := &ChannelResp{}
			json.Unmarshal(resp, ch)
			So(reflect.DeepEqual(ch, expResp), ShouldBeTrue)
		})

		f = frisby.Create("get connection counts").Get(ConnectionCountsPath()).Send()
		f.ExpectStatus(http.StatusOK).AfterContent(func(F *frisby.Frisby, resp []byte, err error) {
			counts := make(map[string]int)
			json.Unmarshal(resp, &counts)
			count, found := counts[expResp.Id]
			So(found, ShouldBeTrue)
			So(count, ShouldEqual, 0)
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

		f = frisby.Create("get connection counts").Get(ConnectionCountsPath()).Send()
		f.ExpectStatus(http.StatusOK).AfterContent(func(F *frisby.Frisby, resp []byte, err error) {
			counts := make(map[string]int)
			json.Unmarshal(resp, &counts)
			_, found := counts[expResp.Id]
			So(found, ShouldBeFalse)
		})

		f = frisby.Create("list channels").Get(ListChannelPath()).Send()

		f.ExpectStatus(http.StatusOK).
			AfterJson(func(F *frisby.Frisby, js *simplejson.Json, err error) {
			So(len(js.MustArray()), ShouldEqual, 0)
		})
	})

	Convey("disalow update channel by removing tags, fields or changing field types", t, func() {
		f := frisby.Create("list channels").Get(ListChannelPath()).Send()

		f.ExpectStatus(http.StatusOK).
			AfterJson(func(F *frisby.Frisby, js *simplejson.Json, err error) {
			So(len(js.MustArray()), ShouldEqual, 0)
		})

		reqBody := Channel{
			Name:            "test",
			Description:     "desc",
			Tags:            []string{"tag1", "tag2"},
			Fields:          map[string]string{"field1": "int"},
			AccessTokens:    []string{"token1"},
			ConnectionLimit: 5,
			MessageRate:     1000,
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

		f = frisby.Create("remove tag").Put(GetChannelPath(chId))
		f.SetJson(map[string][]string{"tags": []string{}}).Send()

		f.ExpectStatus(http.StatusBadRequest).AfterContent(func(F *frisby.Frisby, resp []byte, err error) {
			So(string(resp), ShouldContainSubstring, "removing a tag is not allowed: tag1")
		})

		f = frisby.Create("remove field").Put(GetChannelPath(chId))
		f.SetJson(map[string]map[string]string{"fields": map[string]string{"field2": "string"}}).Send()

		f.ExpectStatus(http.StatusBadRequest).AfterContent(func(F *frisby.Frisby, resp []byte, err error) {
			So(string(resp), ShouldContainSubstring, "removing a field is not allowed: field1")
		})

		f = frisby.Create("update field").Put(GetChannelPath(chId))
		f.SetJson(map[string]map[string]string{"fields": map[string]string{"field1": "string"}}).Send()

		f.ExpectStatus(http.StatusBadRequest).AfterContent(func(F *frisby.Frisby, resp []byte, err error) {
			So(string(resp), ShouldContainSubstring, "changing a field type is not allowed: field1")
		})

		f = frisby.Create("add tag or field").Put(GetChannelPath(chId))
		f.SetJson(map[string]interface{}{
			"fields": map[string]string{"field1": "int", "field2": "boolean"},
			"tags":   []string{"tag1", "tag2", "tag3"},
		}).Send()

		f.ExpectStatus(http.StatusOK)

		expResp := &ChannelResp{
			Id:              chId,
			Name:            reqBody.Name,
			Description:     reqBody.Description,
			Tags:            []string{"tag1", "tag2", "tag3"},
			Fields:          map[string]string{"field1": "int", "field2": "boolean"},
			AccessTokens:    reqBody.AccessTokens,
			ConnectionLimit: reqBody.ConnectionLimit,
			MessageRate:     reqBody.MessageRate,
		}

		f = frisby.Create("get channel").Get(GetChannelPath(chId)).Send()
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
