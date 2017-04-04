package configs

import (
	"bytes"
	. "github.com/eywa/utils"
	. "github.com/smartystreets/goconvey/convey"
	"io/ioutil"
	"log"
	"os"
	"reflect"
	"testing"
	"text/template"
	"time"
)

func TestConfig(t *testing.T) {
	Convey("overwrite default configs", t, func() {
		tmpfile, err := ioutil.TempFile("", "test_config")
		if err != nil {
			log.Fatalln(err.Error())
		}
		defer os.Remove(tmpfile.Name())

		content := `
service:
  host: anotherhost
  device_port: 9091
security:
  ssl:
    cert_file: test_cert_file
  dashboard:
    aes:
      key: aes_key
connections:
  websocket:
    timeouts:
      write: 5s
indices:
  disable: true
  ttl_enabled:
`
		if _, err := tmpfile.WriteString(content); err != nil {
			log.Fatalln(err.Error())
		}
		if err = tmpfile.Close(); err != nil {
			log.Fatalln(err.Error())
		}
		f := tmpfile.Name()
		p := map[string]string{"eywa_home": "path/to/eywa"}

		err = InitializeConfig(f, p)
		So(err, ShouldBeNil)

		buf := bytes.NewBuffer([]byte{})
		t, err := template.New("defaults").Parse(DefaultConfigs)
		So(err, ShouldBeNil)
		err = t.Execute(buf, p)
		So(err, ShouldBeNil)

		expConf, err := ReadConfig(buf)
		if err != nil {
			log.Fatalln(err.Error())
		}
		expConf.Service.Host = "anotherhost"
		expConf.Service.DevicePort = 9091
		expConf.Security.SSL.CertFile = "test_cert_file"
		expConf.Security.Dashboard.AES.KEY = "aes_key"
		expConf.Connections.Websocket.Timeouts.Write = &JSONDuration{5 * time.Second}
		expConf.Indices.Disable = true

		So(reflect.DeepEqual(expConf, Config()), ShouldBeTrue)
	})

	Convey("update default configs", t, func() {
		tmpfile, err := ioutil.TempFile("", "test_config")
		if err != nil {
			log.Fatalln(err.Error())
		}
		defer os.Remove(tmpfile.Name())

		f := tmpfile.Name()
		p := map[string]string{"eywa_home": "path/to/eywa"}

		err = InitializeConfig(f, p)

		expConf, err := Config().DeepCopy()
		So(err, ShouldBeNil)

		settings := map[string]interface{}{
			"security": map[string]interface{}{
				"dashboard": map[string]interface{}{
					"username":     "root1",
					"password":     "cookiecats",
					"token_expiry": "1h",
				},
				"api_key": "new_key",
			},
			"connections": map[string]interface{}{
				"websocket": map[string]interface{}{
					"request_queue_size": 22,
					"timeouts": map[string]interface{}{
						"write":    "40s",
						"read":     "120s",
						"request":  "60s",
						"response": "240s",
					},
					"buffer_sizes": map[string]interface{}{
						"read":  20480,
						"write": 40960,
					},
				},
			},
			"indices": map[string]interface{}{
				"disable": true,
			},
		}
		err = Update(settings)
		So(err, ShouldBeNil)

		expConf.Security.Dashboard.Username = "root1"
		expConf.Security.Dashboard.Password = "cookiecats"
		expConf.Security.Dashboard.TokenExpiry = &JSONDuration{1 * time.Hour}
		expConf.Security.ApiKey = "new_key"
		expConf.Connections.Websocket.RequestQueueSize = 22
		expConf.Connections.Websocket.Timeouts.Write = &JSONDuration{40 * time.Second}
		expConf.Connections.Websocket.Timeouts.Read = &JSONDuration{120 * time.Second}
		expConf.Connections.Websocket.Timeouts.Request = &JSONDuration{60 * time.Second}
		expConf.Connections.Websocket.Timeouts.Response = &JSONDuration{240 * time.Second}
		expConf.Connections.Websocket.BufferSizes.Read = 20480
		expConf.Connections.Websocket.BufferSizes.Write = 40960
		expConf.Indices.Disable = true

		So(reflect.DeepEqual(expConf, Config()), ShouldBeTrue)
		expConf.Indices = nil
		So(reflect.DeepEqual(expConf, Config()), ShouldBeFalse)
	})

	Convey("deep copy returns identical conf", t, func() {
		tmpfile, err := ioutil.TempFile("", "test_config")
		if err != nil {
			log.Fatalln(err.Error())
		}
		defer os.Remove(tmpfile.Name())

		f := tmpfile.Name()
		p := map[string]string{"eywa_home": "path/to/eywa"}

		err = InitializeConfig(f, p)

		So(err, ShouldBeNil)
		_conf, _ := Config().DeepCopy()
		So(reflect.DeepEqual(Config(), _conf), ShouldBeTrue)

		_conf.Indices = nil
		So(reflect.DeepEqual(Config(), _conf), ShouldBeFalse)
	})
}
