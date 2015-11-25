package models

import (
	. "github.com/smartystreets/goconvey/convey"
	. "github.com/vivowares/octopus/connections"
	"reflect"
	"testing"
	"time"
)

type TestConn struct{}

func (tc *TestConn) Host() string                               { return "" }
func (tc *TestConn) IsLocal() bool                              { return true }
func (tc *TestConn) Identifier() string                         { return "test conn 1" }
func (tc *TestConn) Closed() bool                               { return true }
func (tc *TestConn) CreatedAt() time.Time                       { return time.Now() }
func (tc *TestConn) LastPingedAt() time.Time                    { return time.Now() }
func (tc *TestConn) SendAsyncRequest(*Message) error            { return nil }
func (tc *TestConn) SendSyncRequest(*Message) (*Message, error) { return nil, nil }
func (tc *TestConn) SendResponse(*Message) error                { return nil }
func (tc *TestConn) Listen(MessageHandler)                      {}
func (tc *TestConn) Close() error                               { return nil }
func (tc *TestConn) SignalClose()                               {}

func TestChannel(t *testing.T) {

	Convey("successfully parse url encoded point", t, func() {
		tc := &TestConn{}
		ch := &Channel{
			Name:   "test",
			Format: "url",
			Tags:   []string{"city", "country", "state"},
			Fields: map[string]string{"temp": "float", "count": "int", "on": "boolean", "color": "string"},
		}

		raw := `city=sf&country=us&state=ca&temp=67.8&count=19&on=false&color=red`
		_, err := ch.NewPoint(tc, raw)
		So(err, ShouldNotBeNil)

		raw = `city=sf&country=us&state=ca&temp=67.8&count=19&on=false&color=red&timestamp=1448094788939`
		p, err := ch.NewPoint(tc, raw)
		So(err, ShouldBeNil)
		So(p.Raw, ShouldEqual, raw)
		So(reflect.DeepEqual(p.Tags, map[string]string{"city": "sf", "country": "us", "state": "ca", "connection_id": tc.Identifier()}), ShouldBeTrue)
		So(p.Timestamp.UnixNano()/int64(time.Millisecond), ShouldEqual, 1448094788939)
		So(p.Fields["temp"], ShouldEqual, 67.8)
		So(p.Fields["count"], ShouldEqual, 19)
		So(p.Fields["on"], ShouldEqual, false)
		So(p.Fields["color"], ShouldEqual, "red")
	})

	Convey("successfully parse json encoded point", t, func() {
		tc := &TestConn{}
		ch := &Channel{
			Name:   "test",
			Format: "json",
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
		_, err := ch.NewPoint(tc, raw)
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
		p, err := ch.NewPoint(tc, raw)
		So(err, ShouldBeNil)
		So(p.Raw, ShouldEqual, raw)
		So(reflect.DeepEqual(p.Tags, map[string]string{"city": "sf", "country": "us", "state": "ca", "connection_id": tc.Identifier()}), ShouldBeTrue)
		So(p.Timestamp.UnixNano()/int64(time.Millisecond), ShouldEqual, 1448094788939)
		So(p.Fields["temp"], ShouldEqual, 67.8)
		So(p.Fields["count"], ShouldEqual, 19)
		So(p.Fields["on"], ShouldEqual, false)
		So(p.Fields["color"], ShouldEqual, "red")
	})

}
