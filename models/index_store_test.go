package models

import (
	"bytes"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/spf13/viper"
	"testing"
)

func TestIndexStore(t *testing.T) {
	viper.SetConfigType("yaml")

	var yaml = []byte(`
    indices:
      store: influxdb
      host: localhost
      port: 8086
      database: test
  `)

	viper.ReadConfig(bytes.NewBuffer(yaml))
	InitializeIndexStore()

	Convey("writes point to index store", t, func() {
		tc := &TestConn{}
		ch := &Channel{
			Name:        "test",
			Format:      "url",
			Description: "This is a test",
			Tags:        []string{"state", "country", "city"},
			Fields:      map[string]string{"temp": "float", "count": "int", "on": "boolean", "color": "string"},
		}
		raw := `city=sf&country=us&state=ca&temp=67.8&count=19&on=false&color=blue&timestamp=1448094788939`
		p, _ := ch.NewPoint(tc, raw)
		err := p.Store()
		So(err, ShouldBeNil)
	})

}
