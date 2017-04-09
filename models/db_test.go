package models

import (
	. "github.com/smartystreets/goconvey/convey"
	. "github.com/eywa/configs"
	"os"
	"path"
	"testing"
)

func TestDb(t *testing.T) {
	pwd, _ := os.Getwd()
	dbFile := path.Join(pwd, "eywa_test.db")

	Convey("initialize database with no error", t, func() {
		SetConfig(&Conf{
			Database: &DbConf{
				DbType: "sqlite3",
				DbFile: dbFile,
			},
			Logging: &LogsConf{
				Database: &LogConf{
					Level: "debug",
				},
			},
		})
		err := InitializeDB()
		So(err, ShouldEqual, nil)
		CloseDB()
	})

	Convey("initialize database with invalid database type", t, func() {
		SetConfig(&Conf{
			Database: &DbConf{
				DbType: "sql",
				DbFile: "",
			},
			Logging: &LogsConf{
				Database: &LogConf{
					Level: "debug",
				},
			},
		})
		err := InitializeDB()
		So(err, ShouldNotEqual, nil)
		CloseDB()
	})

	os.Remove(dbFile)
}
