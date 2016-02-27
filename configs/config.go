package configs

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/vivowares/eywa/Godeps/_workspace/src/github.com/imdario/mergo"
	"github.com/vivowares/eywa/Godeps/_workspace/src/github.com/spf13/viper"
	"github.com/vivowares/eywa/Godeps/_workspace/src/gopkg.in/yaml.v2"
	. "github.com/vivowares/eywa/utils"
	"io"
	"strings"
	"sync/atomic"
	"text/template"
	"time"
	"unsafe"
)

var cfgPtr unsafe.Pointer
var filename string
var params map[string]string

var DynamicSettings = []string{
	"security.dashboard.username",
	"security.dashboard.password",
	"security.dashboard.token_expiry",
	"websocket_connections.request_queue_size",
	"websocket_connections.timeouts.write",
	"websocket_connections.timeouts.read",
	"websocket_connections.timeouts.request",
	"websocket_connections.timeouts.response",
	"websocket_connections.buffer_sizes.read",
	"websocket_connections.buffer_sizes.write",
}

func Config() *Conf {
	return (*Conf)(cfgPtr)
}

func SetConfig(cfg *Conf) {
	atomic.StorePointer(&cfgPtr, unsafe.Pointer(cfg))
}

func ReadConfig(buf io.Reader) (*Conf, error) {
	v := viper.New()
	v.SetConfigType("yml")
	err := v.ReadConfig(buf)
	if err != nil {
		return nil, err
	}

	serviceConfig := &ServiceConf{
		Host:       v.GetString("service.host"),
		ApiPort:    v.GetInt("service.api_port"),
		DevicePort: v.GetInt("service.device_port"),
		PidFile:    v.GetString("service.pid_file"),
	}

	securityConfig := &SecurityConf{
		Dashboard: &DashboardSecurityConf{
			Username:    v.GetString("security.dashboard.username"),
			Password:    v.GetString("security.dashboard.password"),
			TokenExpiry: v.GetDuration("security.dashboard.token_expiry"),
			AES: &AESConf{
				KEY: v.GetString("security.dashboard.aes.key"),
				IV:  v.GetString("security.dashboard.aes.iv"),
			},
		},
		SSL: &SSLConf{
			CertFile: v.GetString("security.ssl.cert_file"),
			KeyFile:  v.GetString("security.ssl.key_file"),
		},
	}

	dbConfig := &DbConf{
		DbType: v.GetString("database.db_type"),
		DbFile: v.GetString("database.db_file"),
	}

	indexConfig := &IndexConf{
		Disable:          v.GetBool("indices.disable"),
		Host:             v.GetString("indices.host"),
		Port:             v.GetInt("indices.port"),
		NumberOfShards:   v.GetInt("indices.number_of_shards"),
		NumberOfReplicas: v.GetInt("indices.number_of_replicas"),
		TTLEnabled:       v.GetBool("indices.ttl_enabled"),
		TTL:              v.GetDuration("indices.ttl"),
	}

	wsConnConfig := &WsConnectionConf{
		Registry:         v.GetString("websocket_connections.registry"),
		NShards:          v.GetInt("websocket_connections.nshards"),
		InitShardSize:    v.GetInt("websocket_connections.init_shard_size"),
		RequestQueueSize: v.GetInt("websocket_connections.request_queue_size"),
		Timeouts: &WsConnectionTimeoutConf{
			Write:    v.GetDuration("websocket_connections.timeouts.write"),
			Read:     v.GetDuration("websocket_connections.timeouts.read"),
			Request:  v.GetDuration("websocket_connections.timeouts.request"),
			Response: v.GetDuration("websocket_connections.timeouts.response"),
		},
		BufferSizes: &WsConnectionBufferSizeConf{
			Write: v.GetInt("websocket_connections.buffer_sizes.write"),
			Read:  v.GetInt("websocket_connections.buffer_sizes.read"),
		},
	}

	logEywa := &LogConf{
		Filename:   v.GetString("logging.eywa.filename"),
		MaxSize:    v.GetInt("logging.eywa.maxsize"),
		MaxAge:     v.GetInt("logging.eywa.maxage"),
		MaxBackups: v.GetInt("logging.eywa.maxbackups"),
		Level:      v.GetString("logging.eywa.level"),
		BufferSize: v.GetInt("logging.eywa.buffer_size"),
	}

	logIndices := &LogConf{
		Filename:   v.GetString("logging.indices.filename"),
		MaxSize:    v.GetInt("logging.indices.maxsize"),
		MaxAge:     v.GetInt("logging.indices.maxage"),
		MaxBackups: v.GetInt("logging.indices.maxbackups"),
		Level:      v.GetString("logging.indices.level"),
		BufferSize: v.GetInt("logging.indices.buffer_size"),
	}

	logDatabase := &LogConf{
		Filename:   v.GetString("logging.database.filename"),
		MaxSize:    v.GetInt("logging.database.maxsize"),
		MaxAge:     v.GetInt("logging.database.maxage"),
		MaxBackups: v.GetInt("logging.database.maxbackups"),
		Level:      v.GetString("logging.database.level"),
		BufferSize: v.GetInt("logging.database.buffer_size"),
	}

	cfg := &Conf{
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

	return cfg, nil
}

func Update(settings map[string]interface{}) error {
	notAllowed := []string{}
	for k, _ := range settings {
		if !StringSliceContains(DynamicSettings, k) {
			notAllowed = append(notAllowed, k)
		}
	}
	if len(notAllowed) > 0 {
		if len(notAllowed) == 1 {
			return errors.New(fmt.Sprintf("setting: %s is not dynamic", notAllowed[0]))
		} else {
			return errors.New(fmt.Sprintf("settings: %s are not dynamic", strings.Join(notAllowed, ",")))
		}
	}

	_cfg, err := Config().DeepCopy()
	if err != nil {
		return err
	}

	p, err := yaml.Marshal(settings)
	if err != nil {
		return err
	}
	cfg, err := ReadConfig(bytes.NewBuffer(p))
	if err != nil {
		return err
	}
	err = mergo.MergeWithOverwrite(_cfg, *cfg)
	if err != nil {
		return err
	}

	SetConfig(_cfg)
	return nil
}

func InitializeConfig(f string, p map[string]string) error {
	filename = f
	params = p

	buf := bytes.NewBuffer([]byte{})
	_, err := buf.WriteString(DefaultConfigs)
	if err != nil {
		return err
	}
	_cfg, err := ReadConfig(buf)
	if err != nil {
		return err
	}

	t, err := template.ParseFiles(filename)
	if err != nil {
		return err
	}

	buf = bytes.NewBuffer([]byte{})
	err = t.Execute(buf, params)
	if err != nil {
		return err
	}
	cfg, err := ReadConfig(buf)
	if err != nil {
		return err
	}

	err = mergo.MergeWithOverwrite(_cfg, *cfg)
	if err != nil {
		return err
	}

	SetConfig(_cfg)
	return nil
}

type Conf struct {
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
