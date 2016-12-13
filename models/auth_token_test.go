package models

import (
	. "github.com/smartystreets/goconvey/convey"
	. "github.com/eywa/configs"
	. "github.com/eywa/utils"
	"reflect"
	"testing"
	"time"
)

func TestAuthToken(t *testing.T) {
	SetConfig(&Conf{
		Security: &SecurityConf{
			Dashboard: &DashboardSecurityConf{
				Username:    "test_user",
				Password:    "test_password",
				TokenExpiry: &JSONDuration{24 * time.Hour},
				AES: &AESConf{
					KEY: "abcdefg123456789",
					IV:  "abcdefg123456789",
				},
			},
		},
	})

	Convey("encrypts/decrypts auth token", t, func() {
		t, e := NewAuthToken("test_user", "test_password")
		So(e, ShouldBeNil)
		h, e := t.Encrypt()
		So(e, ShouldBeNil)
		_t, e := DecryptAuthToken(h)
		So(e, ShouldBeNil)
		So(reflect.DeepEqual(t, _t), ShouldBeTrue)
	})
}
