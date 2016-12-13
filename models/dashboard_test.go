package models

import (
	. "github.com/smartystreets/goconvey/convey"
	. "github.com/eywa/configs"
	"log"
	"os"
	"path"
	"testing"
)

func TestDashboard(t *testing.T) {
	pwd, _ := os.Getwd()
	dbFile := path.Join(pwd, "eywa_test.db")

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

	InitializeDB()
	DB.LogMode(true)
	DB.SetLogger(log.New(os.Stdout, "", log.LstdFlags))
	DB.AutoMigrate(&Dashboard{})

	Convey("creates/updates/deletes dashboard", t, func() {
		d := &Dashboard{
			Name:        "test",
			Description: "desc",
			Definition:  "definition",
		}

		d.Create()
		var count int
		DB.Model(&Dashboard{}).Count(&count)
		So(count, ShouldEqual, 1)

		d.Name = "updated test"
		d.Update()

		_d := &Dashboard{}
		DB.Model(&Dashboard{}).First(_d)
		So(_d.Name, ShouldEqual, "updated test")

		d.Delete()
		DB.Model(&Dashboard{}).Count(&count)
		So(count, ShouldEqual, 0)
	})

	Convey("validates dashboard before saving", t, func() {
		d := &Dashboard{
			Name:        "",
			Description: "desc",
			Definition:  "def",
		}
		err := d.Create()
		So(err.Error(), ShouldContainSubstring, "name is empty")

		d.Name = "test"
		d.Description = ""
		err = d.Create()
		So(err.Error(), ShouldContainSubstring, "description is empty")
	})

	CloseDB()
	os.Remove(dbFile)
}
