// +build integration

package api_tests

import (
	"encoding/json"
	"fmt"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/verdverm/frisby"
	. "github.com/eywa/configs"
	. "github.com/eywa/models"
	. "github.com/eywa/utils"
	"net/http"
	"reflect"
	"testing"
)

func ConfigsPath() string {
	return fmt.Sprintf("%s/%s", ApiServer, "admin/configs")
}

func TestAdminConfigs(t *testing.T) {

	frisby.Global.SetHeader("Content-Type", "application/json").
		SetHeader("Accept", "application/json")

	Convey("successfully fetches configs", t, func() {
		f := frisby.Create("get configs").SetHeader("Authentication", authStr()).
			Get(ConfigsPath()).Send()

		conf := &Conf{}
		f.ExpectStatus(http.StatusOK).
			AfterContent(func(F *frisby.Frisby, js []byte, err error) {
			json.Unmarshal(js, conf)
		})
		expConf, _ := Config().DeepCopy()
		expConf.Security.SSL = nil
		expConf.Security.Dashboard.AES = nil
		expConf.Security.Dashboard.Password = ""
		expConf.Security.ApiKey = ""
		expConf.Service.PidFile = ""
		expConf.Service.Assets = ""

		So(reflect.DeepEqual(expConf, conf), ShouldBeTrue)
	})

	Convey("successfully updates configs", t, func() {
		newUser := "cookiecats"
		f := frisby.Create("update configs").SetHeader("Authentication", authStr()).
			Put(ConfigsPath()).SetJson(map[string]interface{}{"security": map[string]interface{}{"dashboard": map[string]interface{}{"username": newUser}}}).Send()

		conf := &Conf{}
		f.ExpectStatus(http.StatusOK).
			AfterContent(func(F *frisby.Frisby, js []byte, err error) {
			json.Unmarshal(js, conf)
		})
		expConf, _ := Config().DeepCopy()
		oldUser := expConf.Security.Dashboard.Username
		expConf.Security.SSL = nil
		expConf.Security.Dashboard.AES = nil
		expConf.Security.Dashboard.Password = ""
		expConf.Security.Dashboard.Username = newUser
		expConf.Security.ApiKey = ""
		expConf.Service.PidFile = ""
		expConf.Service.Assets = ""

		So(reflect.DeepEqual(expConf, conf), ShouldBeTrue)

		auth, err := NewAuthToken(
			newUser,
			Config().Security.Dashboard.Password,
		)
		FatalIfErr(err)

		str, err := auth.Encrypt()
		FatalIfErr(err)

		f = frisby.Create("revert configs").SetHeader("Authentication", str).
			Put(ConfigsPath()).SetJson(map[string]interface{}{"security": map[string]interface{}{"dashboard": map[string]interface{}{"username": oldUser}}}).Send()
		f.ExpectStatus(http.StatusOK)
	})

	frisby.Global.PrintReport()
}
