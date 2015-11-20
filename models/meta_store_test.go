package models

import (
	"bytes"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/spf13/viper"
	"reflect"
	"testing"
)

func TestMetaStore(t *testing.T) {
	viper.SetConfigType("yaml")

	var yaml = []byte(`
    persistence:
      store: bolt
      db_file: octopus.db
  `)

	viper.ReadConfig(bytes.NewBuffer(yaml))
	InitializeMetaStore()

	Convey("inserts/finds/deletes the channel successfully", t, func() {
		_, found := MStore.FindChannelByName("test")
		So(found, ShouldBeFalse)

		ch := &Channel{
			Name:        "test",
			Description: "This is a test",
			Tags:        []string{"lat", "long", "country", "city"},
			Fields:      map[string]string{"temp": "float", "order": "int"},
		}
		err := MStore.InsertChannel(ch)
		So(err, ShouldBeNil)

		_ch, found := MStore.FindChannelByName("test")
		So(found, ShouldBeTrue)
		So(reflect.DeepEqual(ch, _ch), ShouldBeTrue)

		MStore.DeleteChannel(_ch)
	})

	Convey("won't update channel if doesn't exist", t, func() {
		_, found := MStore.FindChannelByName("test")
		So(found, ShouldBeFalse)

		ch := &Channel{
			Name:        "test",
			Description: "This is a test",
			Tags:        []string{"lat", "long", "country", "city"},
			Fields:      map[string]string{"temp": "float", "order": "int"},
		}
		err := MStore.UpdateChannel(ch)
		So(err, ShouldNotBeNil)
	})

	Convey("finds a list of channels", t, func() {
		chs, err := MStore.FindChannels()
		So(err, ShouldBeNil)
		So(len(chs), ShouldEqual, 0)

		ch1 := &Channel{
			Name:        "test1",
			Description: "This is test1",
			Tags:        []string{"lat", "long", "country", "city"},
			Fields:      map[string]string{"temp": "float", "order": "int"},
		}
		MStore.InsertChannel(ch1)
		ch2 := &Channel{
			Name:        "test2",
			Description: "This is test2",
			Tags:        []string{"lat", "long", "country", "city"},
			Fields:      map[string]string{"temp": "float", "order": "int"},
		}
		MStore.InsertChannel(ch2)

		chs, err = MStore.FindChannels()
		So(len(chs), ShouldEqual, 2)
		So(reflect.DeepEqual([]*Channel{ch1, ch2}, chs), ShouldBeTrue)

		MStore.DeleteChannel(ch1)
		MStore.DeleteChannel(ch2)
	})

	CloseMetaStore()
}
