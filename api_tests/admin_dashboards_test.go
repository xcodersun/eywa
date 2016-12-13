// +build integration

package api_tests

import (
	"encoding/json"
	"fmt"
	"github.com/bitly/go-simplejson"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/verdverm/frisby"
	. "github.com/eywa/models"
	"log"
	"net/http"
	"os"
	"reflect"
	"testing"
)

type DashboardResp struct {
	Id          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Definition  string `json:"definition"`
}

func ListDashboardPath() string {
	return fmt.Sprintf("%s/%s", ApiServer, "admin/dashboards")
}

func GetDashboardPath(id int) string {
	return fmt.Sprintf("%s/%s/%d", ApiServer, "admin/dashboards", id)
}

func TestAdminDashboards(t *testing.T) {

	InitializeDB()
	DB.LogMode(true)
	DB.SetLogger(log.New(os.Stdout, "", log.LstdFlags))
	DB.DropTableIfExists(&Dashboard{})
	DB.AutoMigrate(&Dashboard{})

	frisby.Global.SetHeader("Content-Type", "application/json").
		SetHeader("Accept", "application/json").
		SetHeader("Authentication", authStr())

	Convey("successfully creates/gets/lists/updates dashboard", t, func() {
		f := frisby.Create("list dashboards").Get(ListDashboardPath()).Send()

		f.ExpectStatus(http.StatusOK).
			AfterJson(func(F *frisby.Frisby, js *simplejson.Json, err error) {
			So(len(js.MustArray()), ShouldEqual, 0)
		})

		reqBody := Dashboard{
			Name:        "test",
			Description: "desc",
			Definition:  "definition",
		}
		f = frisby.Create("create dashboard").Post(ListDashboardPath())
		f.SetJson(reqBody).Send()

		f.ExpectStatus(http.StatusCreated)

		f = frisby.Create("list dashboards").Get(ListDashboardPath()).Send()

		var chId int
		f.ExpectStatus(http.StatusOK).
			AfterJson(func(F *frisby.Frisby, js *simplejson.Json, err error) {
			So(len(js.MustArray()), ShouldEqual, 1)
			ch := js.MustArray()[0].(map[string]interface{})
			_chId, _ := ch["id"].(json.Number).Int64()
			chId = int(_chId)
		})

		expResp := &DashboardResp{
			Id:          chId,
			Name:        reqBody.Name,
			Description: reqBody.Description,
			Definition:  reqBody.Definition,
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
