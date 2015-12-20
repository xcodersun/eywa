// +build integration

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
var ConfigFile string

type DashboardResp struct {
	Id           string            `json:"id"`
	Name         string            `json:"name"`
	Description  string            `json:"description"`
	Definition   string            `json:"definition"`
}

func init() {
	pwd, err := os.Getwd()
	PanicIfErr(err)
	ConfigFile = path.Join(path.Dir(pwd), "configs", "octopus_test.yml")
	PanicIfErr(InitializeConfig(ConfigFile))

	ApiServer = "http://" + Config.Service.Host + ":" + strconv.Itoa(Config.Service.HttpPort)
}

func ListDashboardPath() string {
	return fmt.Sprintf("%s/%s", ApiServer, "channels")
}

func GetDashboardPath(base64Id string) string {
	return fmt.Sprintf("%s/%s/%s", ApiServer, "channels", base64Id)
}

func TestApiDashboards(t *testing.T) {

	InitializeDB()
	DB.DropTableIfExists(&Dashboard{})
	DB.AutoMigrate(&Dashboard{})

	frisby.Global.SetHeader("Content-Type", "application/json").SetHeader("Accept", "application/json")

	Convey("successfully creates/gets/lists/updates dashboard", t, func() {
		f := frisby.Create("list dashboards").Get(ListDashboardPath()).Send()

		f.ExpectStatus(http.StatusOK).
			AfterJson(func(F *frisby.Frisby, js *simplejson.Json, err error) {
			So(len(js.MustArray()), ShouldEqual, 0)
		})

		reqBody := Dashboard{
			Name:         "test",
			Description:  "desc",
			Definition:   "definition",
		}
		f = frisby.Create("create dashboard").Post(ListDashboardPath())
		f.SetJson(reqBody).Send()

		f.ExpectStatus(http.StatusCreated)

		f = frisby.Create("list dashboards").Get(ListDashboardPath()).Send()

		var chId string
		f.ExpectStatus(http.StatusOK).
			AfterJson(func(F *frisby.Frisby, js *simplejson.Json, err error) {
			So(len(js.MustArray()), ShouldEqual, 1)
			ch := js.MustArray()[0].(map[string]interface{})
			chId, _ = ch["id"].(string)
		})

		expResp := &DashboardResp{
			Id:           chId,
			Name:         reqBody.Name,
			Description:  reqBody.Description,
			Definition:   reqBody.Definition,
		}

		f = frisby.Create("get dashboard").Get(GetDashboardPath(chId)).Send()
		f.ExpectStatus(http.StatusOK).AfterContent(func(F *frisby.Frisby, resp []byte, err error) {
			ch := &DashboardResp{}
			json.Unmarshal(resp, ch)
			So(reflect.DeepEqual(ch, expResp), ShouldBeTrue)
		})

		f = frisby.Create("update dashboard").Put(GetDashboardPath(chId))
		f.SetJson(map[string]string{"name": "updated name"}).Send()

		f.ExpectStatus(http.StatusOK)

		f = frisby.Create("get dashboard").Get(GetDashboardPath(chId)).Send()
		expResp.Name = "updated name"
		f.ExpectStatus(http.StatusOK).AfterContent(func(F *frisby.Frisby, resp []byte, err error) {
			ch := &DashboardResp{}
			json.Unmarshal(resp, ch)
			So(reflect.DeepEqual(ch, expResp), ShouldBeTrue)
		})

		f = frisby.Create("delete dashboard").Delete(GetDashboardPath(chId)).Send()
		f.ExpectStatus(http.StatusOK)

		f = frisby.Create("list dashboards").Get(ListDashboardPath()).Send()

		f.ExpectStatus(http.StatusOK).
			AfterJson(func(F *frisby.Frisby, js *simplejson.Json, err error) {
			So(len(js.MustArray()), ShouldEqual, 0)
		})
	})

	frisby.Global.PrintReport()
}
