package configs

import (
	"bytes"
	. "github.com/vivowares/eywa/Godeps/_workspace/src/github.com/smartystreets/goconvey/convey"
	. "github.com/vivowares/eywa/utils"
	"io/ioutil"
	"log"
	"os"
	"reflect"
	"testing"
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
      websocket_connections:
        timeouts:
          write: 5s
          request:
      indices:
        disable: true
        host:
        port:
        ttl_enabled:
    `

		if _, err := tmpfile.WriteString(content); err != nil {
			log.Fatalln(err.Error())
		}
		if err = tmpfile.Close(); err != nil {
			log.Fatalln(err.Error())
		}
		f := tmpfile.Name()
		p := map[string]string{}

		err = InitializeConfig(f, p)

		expConf, err := ReadConfig(bytes.NewBuffer([]byte(DefaultConfigs)))
		if err != nil {
			log.Fatalln(err.Error())
		}
		expConf.Service.Host = "anotherhost"
		expConf.Service.DevicePort = 9091
		expConf.Security.SSL.CertFile = "test_cert_file"
		expConf.Security.Dashboard.AES.KEY = "aes_key"
		expConf.WebSocketConnections.Timeouts.Write = &JSONDuration{5 * time.Second}
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
		p := map[string]string{}

		err = InitializeConfig(f, p)
		oldConfPtr := Config()
		settings := map[string]interface{}{
			"security.dashboard.username":              "root1",
			"security.dashboard.password":              "cookiecats",
			"security.dashboard.token_expiry":          "12h",
			"websocket_connections.request_queue_size": 16,
			"websocket_connections.timeouts.write":     "4s",
			"websocket_connections.timeouts.read":      "12s",
			"websocket_connections.timeouts.request":   "6s",
			"websocket_connections.timeouts.response":  "24s",
			"websocket_connections.buffer_sizes.read":  2048,
			"websocket_connections.buffer_sizes.write": 4096,
		}
		if err = Update(settings); err != nil {
			log.Fatalln(err.Error())
		}

		expConf, err := ReadConfig(bytes.NewBuffer([]byte(DefaultConfigs)))
		if err != nil {
			log.Fatalln(err.Error())
		}
		expConf.Security.Dashboard.Username = "root1"
		expConf.Security.Dashboard.Password = "cookiecats"
		expConf.Security.Dashboard.TokenExpiry = &JSONDuration{12 * time.Hour}
		expConf.WebSocketConnections.RequestQueueSize = 16
		expConf.WebSocketConnections.Timeouts.Write = &JSONDuration{4 * time.Second}
		expConf.WebSocketConnections.Timeouts.Read = &JSONDuration{12 * time.Second}
		expConf.WebSocketConnections.Timeouts.Request = &JSONDuration{6 * time.Second}
		expConf.WebSocketConnections.Timeouts.Response = &JSONDuration{24 * time.Second}
		expConf.WebSocketConnections.BufferSizes.Read = 2048
		expConf.WebSocketConnections.BufferSizes.Write = 4096

		So(reflect.DeepEqual(expConf, Config()), ShouldBeTrue)
		So(reflect.DeepEqual(expConf, oldConfPtr), ShouldBeFalse)

		settings = map[string]interface{}{
			"service.host":                             "localhost",
			"security.dashboard.username":              "root1",
			"security.dashboard.password":              "cookiecats",
			"security.dashboard.token_expiry":          "12h",
			"websocket_connections.request_queue_size": 16,
			"websocket_connections.timeouts.write":     "4s",
			"websocket_connections.timeouts.read":      "12s",
			"websocket_connections.timeouts.request":   "6s",
			"websocket_connections.timeouts.response":  "24s",
			"websocket_connections.buffer_sizes.read":  2048,
			"websocket_connections.buffer_sizes.write": 4096,
		}
		err = Update(settings)
		So(err, ShouldNotBeNil)
	})

	Convey("deep copy returns identical conf", t, func() {
		tmpfile, err := ioutil.TempFile("", "test_config")
		if err != nil {
			log.Fatalln(err.Error())
		}
		defer os.Remove(tmpfile.Name())

		f := tmpfile.Name()
		p := map[string]string{}

		err = InitializeConfig(f, p)

		So(err, ShouldBeNil)
		_conf, _ := Config().DeepCopy()
		So(reflect.DeepEqual(Config(), _conf), ShouldBeTrue)
	})
}
