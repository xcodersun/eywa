package configs

import (
	"github.com/spf13/viper"
	"time"
)

var Config *Conf

func InitializeConfig(configPath string) error {
	viper.SetConfigName("octopus")
	viper.AddConfigPath(configPath)
	err := viper.ReadInConfig()
	if err != nil {
		return err
	}

	serviceConfig := &ServiceConf{
		Host:     viper.GetString("service.host"),
		HttpPort: viper.GetInt("service.http_port"),
		WsPort:   viper.GetInt("service.ws_port"),
	}

	dbConfig := &DbConf{
		DbType: viper.GetString("database.db_type"),
		DbFile: viper.GetString("database.db_file"),
	}

	indexConfig := &IndexConf{
		Host:   viper.GetString("indices.host"),
		Port:   viper.GetInt("indices.port"),
		DbName: viper.GetString("indices.db_name"),
	}

	connConfig := &ConnectionConf{
		Store:  viper.GetString("connections.store"),
		Expiry: viper.GetDuration("connections.expiry"),
		Timeouts: &ConnectionTimeoutConf{
			Write:    viper.GetDuration("connections.timeouts.write"),
			Read:     viper.GetDuration("connections.timeouts.read"),
			Response: viper.GetDuration("connections.timeouts.response"),
		},
		BufferSizes: &ConnectionBufferSizeConf{
			Write: viper.GetInt("connections.buffer_sizes.write"),
			Read:  viper.GetInt("connections.buffer_sizes.read"),
		},
	}

	Config = &Conf{
		Service:     serviceConfig,
		Connections: connConfig,
		Indices:     indexConfig,
		Database:    dbConfig,
	}

	return nil
}

type Conf struct {
	Service     *ServiceConf
	Connections *ConnectionConf
	Indices     *IndexConf
	Database    *DbConf
}

type DbConf struct {
	DbType string
	DbFile string
}

type IndexConf struct {
	Host   string
	Port   int
	DbName string
}

type ServiceConf struct {
	Host     string
	HttpPort int
	WsPort   int
}

type ConnectionConf struct {
	Store       string
	Expiry      time.Duration
	Timeouts    *ConnectionTimeoutConf
	BufferSizes *ConnectionBufferSizeConf
}

type ConnectionTimeoutConf struct {
	Write    time.Duration
	Read     time.Duration
	Response time.Duration
}

type ConnectionBufferSizeConf struct {
	Write int
	Read  int
}
