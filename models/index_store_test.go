package models

import (
	"bytes"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/spf13/viper"
	// "reflect"
	"fmt"
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
		ch := &Channel{
			Name:        "test",
			Description: "This is a test",
			Tags:        []string{"state", "country", "city"},
			Fields:      map[string]string{"temp": "float", "count": "int", "on": "boolean", "color": "string"},
		}
		raw := `city=sf&country=us&state=ca&temp=67.8&count=19&on=false&color=red&timestamp=1448094788939`
		p, _ := ch.NewPoint("url", raw)
		fmt.Println("%+v", p)
		err := p.Store()
		So(err, ShouldBeNil)
	})
}
