// +build integration

package api_tests

import (
	"encoding/json"
	"fmt"
	"github.com/vivowares/eywa/Godeps/_workspace/src/github.com/bitly/go-simplejson"
	. "github.com/vivowares/eywa/Godeps/_workspace/src/github.com/smartystreets/goconvey/convey"
	"github.com/vivowares/eywa/Godeps/_workspace/src/github.com/verdverm/frisby"
	. "github.com/vivowares/eywa/configs"
	. "github.com/vivowares/eywa/utils"
	"net/http"
	"testing"
	"time"
)

func UserLoginPath() string {
	return fmt.Sprintf("%s/%s", ApiServer, "login")
}

func TestAuthLogin(t *testing.T) {

	frisby.Global.SetHeader("Content-Type", "application/json").
		SetHeader("Accept", "application/json")

	Convey("successfully login the user and get the auth token", t, func() {
		f := frisby.Create("user login").
			BasicAuth(Config().Security.Dashboard.Username, Config().Security.Dashboard.Password).
			Get(UserLoginPath()).Send()

		f.ExpectStatus(http.StatusOK).
			AfterJson(func(F *frisby.Frisby, js *simplejson.Json, err error) {
			So(len(js.MustMap()["auth_token"].(string)), ShouldBeGreaterThan, 0)
			ts, _ := js.MustMap()["expires_at"].(json.Number).Int64()
			exp := time.Unix(MilliSecToSec(ts), MilliSecToNano(ts))
			So(exp.After(time.Now().Add(-1*time.Minute).Add(Config().Security.Dashboard.TokenExpiry)), ShouldBeTrue)
		})
	})

	Convey("refused access without auth token", t, func() {
		f := frisby.Create("list channels").Get(ListChannelPath()).Send()

		f.ExpectStatus(http.StatusUnauthorized)
	})

	frisby.Global.PrintReport()
}
