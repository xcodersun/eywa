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

	logConfig := &LogsConf{
		AccessLog: &LogConf{
			Filename:   viper.GetString("logs.access.filename"),
			MaxSize:    viper.GetInt("logs.access.maxsize"),
			MaxAge:     viper.GetInt("logs.access.maxage"),
			MaxBackups: viper.GetInt("logs.access.maxbackups"),
		},
		ConnectionLog: &LogConf{
			Filename:   viper.GetString("logs.connection.filename"),
			MaxSize:    viper.GetInt("logs.connection.maxsize"),
			MaxAge:     viper.GetInt("logs.connection.maxage"),
			MaxBackups: viper.GetInt("logs.connection.maxbackups"),
		},
	}

	Config = &Conf{
		Service:     serviceConfig,
		Connections: connConfig,
		Indices:     indexConfig,
		Database:    dbConfig,
		Logs:        logConfig,
	}

	return nil
}

type Conf struct {
	Service     *ServiceConf
	Connections *ConnectionConf
	Indices     *IndexConf
	Database    *DbConf
	Logs        *LogsConf
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

type LogsConf struct {
	AccessLog     *LogConf
	ConnectionLog *LogConf
}

type LogConf struct {
	Filename   string
	MaxSize    int
	MaxAge     int
	MaxBackups int
}
