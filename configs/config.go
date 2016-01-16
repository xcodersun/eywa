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
	}

	dbConfig := &DbConf{
		DbType:  viper.GetString("database.db_type"),
		DbFile:  viper.GetString("database.db_file"),
		Logging: viper.GetBool("database.logging"),
	}

	indexConfig := &IndexConf{
		Host:             viper.GetString("indices.host"),
		Port:             viper.GetInt("indices.port"),
		NumberOfShards:   viper.GetInt("indices.number_of_shards"),
		NumberOfReplicas: viper.GetInt("indices.number_of_replicas"),
		TTLEnabled:       viper.GetBool("indices.ttl_enabled"),
		TTL:              viper.GetDuration("indices.ttl"),
	}

	connConfig := &ConnectionConf{
		Registry:         viper.GetString("connections.registry"),
		NShards:          viper.GetInt("connections.nshards"),
		InitShardSize:    viper.GetInt("connections.init_shard_size"),
		RequestQueueSize: viper.GetInt("connections.request_queue_size"),
		Expiry:           viper.GetDuration("connections.expiry"),
		Timeouts: &ConnectionTimeoutConf{
			Write:    viper.GetDuration("connections.timeouts.write"),
			Read:     viper.GetDuration("connections.timeouts.read"),
			Request:  viper.GetDuration("connections.timeouts.request"),
			Response: viper.GetDuration("connections.timeouts.response"),
		},
		BufferSizes: &ConnectionBufferSizeConf{
			Write: viper.GetInt("connections.buffer_sizes.write"),
			Read:  viper.GetInt("connections.buffer_sizes.read"),
		},
	}

	logConfig := &LogConf{
		Filename:   viper.GetString("logging.filename"),
		MaxSize:    viper.GetInt("logging.maxsize"),
		MaxAge:     viper.GetInt("logging.maxage"),
		MaxBackups: viper.GetInt("logging.maxbackups"),
		Level:      viper.GetString("logging.level"),
		BufferSize: viper.GetInt("logging.buffer_size"),
	}

	cfg := &Conf{
		Service:     serviceConfig,
		Security:    securityConfig,
		Connections: connConfig,
		Indices:     indexConfig,
		Database:    dbConfig,
		Logging:     logConfig,
	}

	atomic.StorePointer(&cfgPtr, unsafe.Pointer(cfg))
	return nil
}

func InitializeConfig(f string) error {
	filename = f
	return Reload()
}

type Conf struct {
	Service     *ServiceConf    `json:"service"`
	Security    *SecurityConf   `json:"security"`
	Connections *ConnectionConf `json:"connections"`
	Indices     *IndexConf      `json:"indices"`
	Database    *DbConf         `json:"database"`
	Logging     *LogConf        `json:"logging"`
}

type DbConf struct {
	DbType  string `json:"db_type"`
	DbFile  string `json:"db_file"`
	Logging bool   `json:"logging"`
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

type ConnectionConf struct {
	Registry         string                    `json:"registry"`
	NShards          int                       `json:"nshards"`
	InitShardSize    int                       `json:"init_shard_size"`
	RequestQueueSize int                       `json:"request_queue_size"`
	Expiry           time.Duration             `json:"expiry"`
	Timeouts         *ConnectionTimeoutConf    `json:"timeouts"`
	BufferSizes      *ConnectionBufferSizeConf `json:"buffer_sizes"`
}

type ConnectionTimeoutConf struct {
	Write    time.Duration `json:"write"`
	Read     time.Duration `json:"read"`
	Request  time.Duration `json:"request"`
	Response time.Duration `json:"response"`
}

type ConnectionBufferSizeConf struct {
	Write int `json:"write"`
	Read  int `json:"read"`
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
