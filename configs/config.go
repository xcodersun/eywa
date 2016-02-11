package configs

import (
	"bytes"
	"errors"
	"github.com/vivowares/octopus/Godeps/_workspace/src/github.com/spf13/viper"
	"os"
	"path/filepath"
	"sync/atomic"
	"text/template"
	"time"
	"unsafe"
)

var cfgPtr unsafe.Pointer
var filename string

func Config() *Conf {
	return (*Conf)(cfgPtr)
}

func SetConfig(cfg *Conf) {
	atomic.StorePointer(&cfgPtr, unsafe.Pointer(cfg))
}

func Reload() error {
	t, err := template.ParseFiles(filename)
	if err != nil {
		return err
	}
	buf := bytes.NewBuffer([]byte{})
	octopus_home := os.Getenv("OCTOPUS_HOME")
	if len(octopus_home) == 0 {
		return errors.New("ENV OCTOPUS_HOME is not set")
	}

	octopus_home, err = filepath.Abs(octopus_home)
	if err != nil {
		return err
	}

	err = t.Execute(buf, map[string]string{"octopus_home": octopus_home})
	if err != nil {
		return err
	}

	viper.SetConfigType("yml")
	viper.ReadConfig(buf)
	if err != nil {
		return err
	}

	serviceConfig := &ServiceConf{
		Host:     viper.GetString("service.host"),
		HttpPort: viper.GetInt("service.http_port"),
		WsPort:   viper.GetInt("service.ws_port"),
		PidFile:  viper.GetString("service.pid_file"),
	}

	securityConfig := &SecurityConf{
		Dashboard: &DashboardSecurityConf{
			Username:    viper.GetString("security.dashboard.username"),
			Password:    viper.GetString("security.dashboard.password"),
			TokenExpiry: viper.GetDuration("security.dashboard.token_expiry"),
			AES: &AESConf{
				KEY: viper.GetString("security.dashboard.aes.key"),
				IV:  viper.GetString("security.dashboard.aes.iv"),
			},
		},
		SSL: &SSLConf{
			CertFile: viper.GetString("security.ssl.certfile"),
			KeyFile:  viper.GetString("security.ssl.keyfile"),
		},
	}

	dbConfig := &DbConf{
		DbType: viper.GetString("database.db_type"),
		DbFile: viper.GetString("database.db_file"),
	}

	indexConfig := &IndexConf{
		Host:             viper.GetString("indices.host"),
		Port:             viper.GetInt("indices.port"),
		NumberOfShards:   viper.GetInt("indices.number_of_shards"),
		NumberOfReplicas: viper.GetInt("indices.number_of_replicas"),
		TTLEnabled:       viper.GetBool("indices.ttl_enabled"),
		TTL:              viper.GetDuration("indices.ttl"),
	}

	wsConnConfig := &WsConnectionConf{
		Registry:         viper.GetString("ws_connections.registry"),
		NShards:          viper.GetInt("ws_connections.nshards"),
		InitShardSize:    viper.GetInt("ws_connections.init_shard_size"),
		RequestQueueSize: viper.GetInt("ws_connections.request_queue_size"),
		Expiry:           viper.GetDuration("ws_connections.expiry"),
		Timeouts: &WsConnectionTimeoutConf{
			Write:    viper.GetDuration("ws_connections.timeouts.write"),
			Read:     viper.GetDuration("ws_connections.timeouts.read"),
			Request:  viper.GetDuration("ws_connections.timeouts.request"),
			Response: viper.GetDuration("ws_connections.timeouts.response"),
		},
		BufferSizes: &WsConnectionBufferSizeConf{
			Write: viper.GetInt("ws_connections.buffer_sizes.write"),
			Read:  viper.GetInt("ws_connections.buffer_sizes.read"),
		},
	}

	logOctopus := &LogConf{
		Filename:   viper.GetString("logging.octopus.filename"),
		MaxSize:    viper.GetInt("logging.octopus.maxsize"),
		MaxAge:     viper.GetInt("logging.octopus.maxage"),
		MaxBackups: viper.GetInt("logging.octopus.maxbackups"),
		Level:      viper.GetString("logging.octopus.level"),
		BufferSize: viper.GetInt("logging.octopus.buffer_size"),
	}

	logIndices := &LogConf{
		Filename:   viper.GetString("logging.indices.filename"),
		MaxSize:    viper.GetInt("logging.indices.maxsize"),
		MaxAge:     viper.GetInt("logging.indices.maxage"),
		MaxBackups: viper.GetInt("logging.indices.maxbackups"),
		Level:      viper.GetString("logging.indices.level"),
		BufferSize: viper.GetInt("logging.indices.buffer_size"),
	}

	logDatabase := &LogConf{
		Filename:   viper.GetString("logging.database.filename"),
		MaxSize:    viper.GetInt("logging.database.maxsize"),
		MaxAge:     viper.GetInt("logging.database.maxage"),
		MaxBackups: viper.GetInt("logging.database.maxbackups"),
		Level:      viper.GetString("logging.database.level"),
		BufferSize: viper.GetInt("logging.database.buffer_size"),
	}

	cfg := &Conf{
		AutoReload:           viper.GetDuration("auto_reload"),
		Service:              serviceConfig,
		Security:             securityConfig,
		WebSocketConnections: wsConnConfig,
		Indices:              indexConfig,
		Database:             dbConfig,
		Logging: &LogsConf{
			Octopus:  logOctopus,
			Indices:  logIndices,
			Database: logDatabase,
		},
	}

	atomic.StorePointer(&cfgPtr, unsafe.Pointer(cfg))
	return nil
}

func InitializeConfig(f string) error {
	filename = f
	err := Reload()
	if err == nil && Config().AutoReload.Nanoseconds() > 0 {
		go func() {
			for {
				time.Sleep(Config().AutoReload)
				Reload()
			}
		}()
	}
	return err
}

type Conf struct {
	AutoReload           time.Duration     `json:"auto_reload"`
	Service              *ServiceConf      `json:"service"`
	Security             *SecurityConf     `json:"security"`
	WebSocketConnections *WsConnectionConf `json:"ws_connections"`
	Indices              *IndexConf        `json:"indices"`
	Database             *DbConf           `json:"database"`
	Logging              *LogsConf         `json:"logging"`
}

type DbConf struct {
	DbType string `json:"db_type"`
	DbFile string `json:"db_file"`
}

type IndexConf struct {
	Host             string        `json:"host"`
	Port             int           `json:"port"`
	NumberOfShards   int           `json:"number_of_shards"`
	NumberOfReplicas int           `json:"number_of_replicas"`
	TTLEnabled       bool          `json:"ttl_enabled"`
	TTL              time.Duration `json:"ttl"`
}

type ServiceConf struct {
	Host     string `json:"host"`
	HttpPort int    `json:"http_port"`
	WsPort   int    `json:"ws_port"`
	PidFile  string `json:"pid_file"`
}

type WsConnectionConf struct {
	Registry         string                      `json:"registry"`
	NShards          int                         `json:"nshards"`
	InitShardSize    int                         `json:"init_shard_size"`
	RequestQueueSize int                         `json:"request_queue_size"`
	Expiry           time.Duration               `json:"expiry"`
	Timeouts         *WsConnectionTimeoutConf    `json:"timeouts"`
	BufferSizes      *WsConnectionBufferSizeConf `json:"buffer_sizes"`
}

type WsConnectionTimeoutConf struct {
	Write    time.Duration `json:"write"`
	Read     time.Duration `json:"read"`
	Request  time.Duration `json:"request"`
	Response time.Duration `json:"response"`
}

type WsConnectionBufferSizeConf struct {
	Write int `json:"write"`
	Read  int `json:"read"`
}

type LogsConf struct {
	Octopus  *LogConf `json:"octopus"`
	Indices  *LogConf `json:"indices"`
	Database *LogConf `json:"database"`
}

type LogConf struct {
	Filename   string `json:"filename"`
	MaxSize    int    `json:"maxsize"`
	MaxAge     int    `json:"maxage"`
	MaxBackups int    `json:"maxbackups"`
	Level      string `json:"level"`
	BufferSize int    `json:"buffer_size"`
}

type SecurityConf struct {
	Dashboard *DashboardSecurityConf `json:"dashboard"`
	SSL       *SSLConf               `json:"ssl"`
}

type DashboardSecurityConf struct {
	Username    string        `json:"username"`
	Password    string        `json:"password"`
	TokenExpiry time.Duration `json:"token_expiry"`
	AES         *AESConf      `json:"aes"`
}

type AESConf struct {
	KEY string `json:"key"`
	IV  string `json:"iv"`
}

type SSLConf struct {
	CertFile string `json:"cert_file"`
	KeyFile  string `json:"cert_key"`
}
