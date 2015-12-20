package models

import (
	. "github.com/vivowares/octopus/Godeps/_workspace/src/github.com/smartystreets/goconvey/convey"
	. "github.com/vivowares/octopus/configs"
	"os"
	"path"
	"strconv"
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

		c.Name = "updated test"
		c.Update()

		_c := &Channel{}
		DB.Model(&Channel{}).First(_c)
		So(_c.Name, ShouldEqual, "updated test")

		c.Delete()
		DB.Model(&Channel{}).Count(&count)
		So(count, ShouldEqual, 0)
	})

	Convey("validates channel before saving", t, func() {
		c := &Channel{
			Name:         "",
			Description:  "desc",
			Tags:         []string{"tag1", "tag2"},
			Fields:       map[string]string{"field1": "int"},
			AccessTokens: []string{"token1"},
		}
		err := c.Create()
		So(err.Error(), ShouldContainSubstring, "name is empty")

		c.Name = "test"
		c.Description = ""
		err = c.Create()
		So(err.Error(), ShouldContainSubstring, "description is empty")

		c.Description = "desc"
		c.AccessTokens = []string{}
		err = c.Create()
		So(err.Error(), ShouldContainSubstring, "access_tokens are empty")

		tags := []string{}
		for i := 0; i < 65; i++ {
			tags = append(tags, "tag"+strconv.Itoa(i))
		}
		c.Tags = tags
		c.AccessTokens = []string{"token1"}
		err = c.Create()
		So(err.Error(), ShouldContainSubstring, "too many tags")

		tags[1] = "@#$Q$!"
		c.Tags = tags[:64]
		err = c.Create()
		So(err.Error(), ShouldContainSubstring, "invalid tag name")

		tags[1] = "tag0"
		c.Tags = tags[:64]
		err = c.Create()
		So(err.Error(), ShouldContainSubstring, "duplicate tag name")

		c.Tags[1] = "tag1"
		c.Fields = map[string]string{}
		err = c.Create()
		So(err.Error(), ShouldContainSubstring, "fields are empty")

		fields := map[string]string{}
		for i := 0; i < 129; i++ {
			fields["field"+strconv.Itoa(i)] = "int"
		}
		c.Fields = fields
		err = c.Create()
		So(err.Error(), ShouldContainSubstring, "too many fields")

		delete(c.Fields, "field128")
		c.Fields["tag0"] = "float"
		err = c.Create()
		So(err.Error(), ShouldContainSubstring, "conflicting tag name")

		delete(c.Fields, "tag0")
		c.Fields["field0"] = "random"
		err = c.Create()
		So(err.Error(), ShouldContainSubstring, "unsupported datatype")

		delete(c.Fields, "field0")
		c.Fields["!@#$!@#"] = "random"
		err = c.Create()
		So(err.Error(), ShouldContainSubstring, "invalid field name")

		var count int
		DB.Model(&Channel{}).Count(&count)
		So(count, ShouldEqual, 0)
	})

	CloseDB()
	os.Remove(dbFile)
}
