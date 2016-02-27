package configs

import (
	"bytes"
	"encoding/json"
	"github.com/vivowares/eywa/Godeps/_workspace/src/github.com/spf13/viper"
	"sync/atomic"
	"text/template"
	"time"
	"unsafe"
)

var cfgPtr unsafe.Pointer
var filename string
var params map[string]string

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
	err = t.Execute(buf, params)
	if err != nil {
		return err
	}

	viper.SetConfigType("yml")
	err = viper.ReadConfig(buf)
	if err != nil {
		return err
	}

	serviceConfig := &ServiceConf{
		Host:       viper.GetString("service.host"),
		ApiPort:    viper.GetInt("service.api_port"),
		DevicePort: viper.GetInt("service.device_port"),
		PidFile:    viper.GetString("service.pid_file"),
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
		Disable:          viper.GetBool("indices.disable"),
		Host:             viper.GetString("indices.host"),
		Port:             viper.GetInt("indices.port"),
		NumberOfShards:   viper.GetInt("indices.number_of_shards"),
		NumberOfReplicas: viper.GetInt("indices.number_of_replicas"),
		TTLEnabled:       viper.GetBool("indices.ttl_enabled"),
		TTL:              viper.GetDuration("indices.ttl"),
	}

	wsConnConfig := &WsConnectionConf{
		Registry:         viper.GetString("websocket_connections.registry"),
		NShards:          viper.GetInt("websocket_connections.nshards"),
		InitShardSize:    viper.GetInt("websocket_connections.init_shard_size"),
		RequestQueueSize: viper.GetInt("websocket_connections.request_queue_size"),
		Expiry:           viper.GetDuration("websocket_connections.expiry"),
		Timeouts: &WsConnectionTimeoutConf{
			Write:    viper.GetDuration("websocket_connections.timeouts.write"),
			Read:     viper.GetDuration("websocket_connections.timeouts.read"),
			Request:  viper.GetDuration("websocket_connections.timeouts.request"),
			Response: viper.GetDuration("websocket_connections.timeouts.response"),
		},
		BufferSizes: &WsConnectionBufferSizeConf{
			Write: viper.GetInt("websocket_connections.buffer_sizes.write"),
			Read:  viper.GetInt("websocket_connections.buffer_sizes.read"),
		},
	}

	logEywa := &LogConf{
		Filename:   viper.GetString("logging.eywa.filename"),
		MaxSize:    viper.GetInt("logging.eywa.maxsize"),
		MaxAge:     viper.GetInt("logging.eywa.maxage"),
		MaxBackups: viper.GetInt("logging.eywa.maxbackups"),
		Level:      viper.GetString("logging.eywa.level"),
		BufferSize: viper.GetInt("logging.eywa.buffer_size"),
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
			Eywa:     logEywa,
			Indices:  logIndices,
			Database: logDatabase,
		},
	}

	SetConfig(cfg)
	return nil
}

func Update(p []byte) error {
	cfg, err := Config().DeepCopy()
	if err != nil {
		return err
	}

	err = json.Unmarshal(p, cfg)
	if err != nil {
		return err
	}

	SetConfig(cfg)
	return nil
}

func InitializeConfig(f string, p map[string]string) error {
	filename = f
	params = p

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
	WebSocketConnections *WsConnectionConf `json:"websocket_connections"`
	Indices              *IndexConf        `json:"indices"`
	Database             *DbConf           `json:"database"`
	Logging              *LogsConf         `json:"logging"`
}

func (cfg *Conf) DeepCopy() (*Conf, error) {
	js, err := json.Marshal(cfg)
	if err != nil {
		return nil, err
	}

	_cfg := &Conf{}
	err = json.Unmarshal(js, _cfg)
	if err != nil {
		return nil, err
	}

	_cfg.Security.Dashboard.Username = cfg.Security.Dashboard.Username
	_cfg.Security.Dashboard.Password = cfg.Security.Dashboard.Password
	_cfg.Security.Dashboard.AES = cfg.Security.Dashboard.AES

	return _cfg, nil
}

type DbConf struct {
	DbType string `json:"db_type"`
	DbFile string `json:"db_file"`
}

type IndexConf struct {
	Disable          bool          `json:"disable"`
	Host             string        `json:"host"`
	Port             int           `json:"port"`
	NumberOfShards   int           `json:"number_of_shards"`
	NumberOfReplicas int           `json:"number_of_replicas"`
	TTLEnabled       bool          `json:"ttl_enabled"`
	TTL              time.Duration `json:"ttl"`
}

type ServiceConf struct {
	Host       string `json:"host"`
	ApiPort    int    `json:"api_port"`
	DevicePort int    `json:"device_port"`
	PidFile    string `json:"pid_file"`
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
	Eywa     *LogConf `json:"eywa"`
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
	Username    string        `json:"-"`
	Password    string        `json:"-"`
	TokenExpiry time.Duration `json:"token_expiry"`
	AES         *AESConf      `json:"-"`
}

type AESConf struct {
	KEY string `json:"-"`
	IV  string `json:"-"`
}

type SSLConf struct {
	CertFile string `json:"cert_file"`
	KeyFile  string `json:"cert_key"`
}
