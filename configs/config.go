package configs

import (
	"bytes"
	"errors"
	"github.com/vivowares/octopus/Godeps/_workspace/src/github.com/spf13/viper"
	"os"
	"path/filepath"
	"text/template"
	"time"
)

var Config *Conf

func InitializeConfig(filename string) error {

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

	Config = &Conf{
		Service:     serviceConfig,
		Security:    securityConfig,
		Connections: connConfig,
		Indices:     indexConfig,
		Database:    dbConfig,
		Logging:     logConfig,
	}

	return nil
}

type Conf struct {
	Service     *ServiceConf
	Security    *SecurityConf
	Connections *ConnectionConf
	Indices     *IndexConf
	Database    *DbConf
	Logging     *LogConf
}

type DbConf struct {
	DbType  string
	DbFile  string
	Logging bool
}

type IndexConf struct {
	Host             string
	Port             int
	NumberOfShards   int
	NumberOfReplicas int
	TTLEnabled       bool
	TTL              time.Duration
}

type ServiceConf struct {
	Host     string
	HttpPort int
	WsPort   int
	PidFile  string
}

type ConnectionConf struct {
	Registry         string
	NShards          int
	InitShardSize    int
	RequestQueueSize int
	Expiry           time.Duration
	Timeouts         *ConnectionTimeoutConf
	BufferSizes      *ConnectionBufferSizeConf
}

type ConnectionTimeoutConf struct {
	Write    time.Duration
	Read     time.Duration
	Request  time.Duration
	Response time.Duration
}

type ConnectionBufferSizeConf struct {
	Write int
	Read  int
}

type LogConf struct {
	Filename   string
	MaxSize    int
	MaxAge     int
	MaxBackups int
	Level      string
	BufferSize int
}

type SecurityConf struct {
	Dashboard *DashboardSecurityConf
}

type DashboardSecurityConf struct {
	Username    string
	Password    string
	TokenExpiry time.Duration
	AES         *AESConf
}

type AESConf struct {
	KEY string
	IV  string
}
