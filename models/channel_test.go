package models

import (
	"github.com/kr/pretty"
	. "github.com/smartystreets/goconvey/convey"
	. "github.com/vivowares/octopus/configs"
	"os"
	"path"
	"testing"
)

func TestChannel(t *testing.T) {
	pwd, _ := os.Getwd()
	dbFile := path.Join(pwd, "octopus_test.db")

	Config = &Conf{
		Database: &DbConf{
			DbType: "sqlite3",
			DbFile: dbFile,
		},
	}

	InitializeDB()
	DB.AutoMigrate(&Channel{})
	DB.LogMode(true)

	Convey("creates/updates/deletes channel", t, func() {
		c := &Channel{
			Name:         "test",
			Description:  "desc",
			Tags:         []string{"tag1", "tag2"},
			Fields:       map[string]string{"field1": "int"},
			AccessTokens: []string{"token1"},
		}

		c.Create()
		var count int
		DB.Model(&Channel{}).Count(&count)
		So(count, ShouldEqual, 1)
		pretty.Println(c)

		c.Name = "updated test"
		c.Update()

		_c := &Channel{}
		DB.Model(&Channel{}).First(_c)
		So(_c.Name, ShouldEqual, "updated test")
		pretty.Println(_c)

		c.Delete()
		DB.Model(&Channel{}).Count(&count)
		So(count, ShouldEqual, 0)
	})

	//TODO
	Convey("validates channel before saving", t, func() {})

	CloseDB()
	os.Remove(dbFile)
}
