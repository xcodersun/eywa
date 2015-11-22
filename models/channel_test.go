package models

import (
	. "github.com/smartystreets/goconvey/convey"
	"reflect"
	"testing"
	"time"
)

func TestChannel(t *testing.T) {

	Convey("successfully parse url encoded point", t, func() {
		ch := &Channel{
			Name:   "test",
			Tags:   []string{"city", "country", "state"},
			Fields: map[string]string{"temp": "float", "count": "int", "on": "boolean", "color": "string"},
		}

		raw := `city=sf&country=us&state=ca&temp=67.8&count=19&on=false&color=red`
		_, err := ch.NewPoint("url", raw)
		So(err, ShouldNotBeNil)

		raw = `city=sf&country=us&state=ca&temp=67.8&count=19&on=false&color=red&timestamp=1448094788939`
		p, err := ch.NewPoint("url", raw)
		So(err, ShouldBeNil)
		So(p.Format, ShouldEqual, "url")
		So(p.Raw, ShouldEqual, raw)
		So(reflect.DeepEqual(p.Tags, map[string]string{"city": "sf", "country": "us", "state": "ca"}), ShouldBeTrue)
		So(p.Timestamp.UnixNano()/int64(time.Millisecond), ShouldEqual, 1448094788939)
		So(p.Fields["temp"], ShouldEqual, 67.8)
		So(p.Fields["count"], ShouldEqual, 19)
		So(p.Fields["on"], ShouldEqual, false)
		So(p.Fields["color"], ShouldEqual, "red")
	})

	Convey("successfully parse json encoded point", t, func() {
		ch := &Channel{
			Name:   "test",
			Tags:   []string{"city", "country", "state"},
			Fields: map[string]string{"temp": "float", "count": "int", "on": "boolean", "color": "string"},
		}

		raw := `
		{
			"city": "sf",
			"country": "us",
			"state": "ca",
			"temp": 67.8,
			"count": 19,
			"on": false,
			"color": "red"
		}
		`
		_, err := ch.NewPoint("url", raw)
		So(err, ShouldNotBeNil)

		raw = `
		{
			"city": "sf",
			"country": "us",
			"state": "ca",
			"temp": 67.8,
			"count": 19,
			"on": false,
			"color": "red",
			"timestamp": 1448094788939
		}
		`
		p, err := ch.NewPoint("json", raw)
		So(err, ShouldBeNil)
		So(p.Format, ShouldEqual, "json")
		So(p.Raw, ShouldEqual, raw)
		So(reflect.DeepEqual(p.Tags, map[string]string{"city": "sf", "country": "us", "state": "ca"}), ShouldBeTrue)
		So(p.Timestamp.UnixNano()/int64(time.Millisecond), ShouldEqual, 1448094788939)
		So(p.Fields["temp"], ShouldEqual, 67.8)
		So(p.Fields["count"], ShouldEqual, 19)
		So(p.Fields["on"], ShouldEqual, false)
		So(p.Fields["color"], ShouldEqual, "red")
	})

}
