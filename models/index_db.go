package models

import (
	"fmt"
	influx "github.com/vivowares/octopus/Godeps/_workspace/src/github.com/influxdb/influxdb/client/v2"
	. "github.com/vivowares/octopus/configs"
	// "net/url"
)

var IndexDB influx.Client

func CloseIndexDB() error {
	return IndexDB.Close()
}

func InitializeIndexDB() error {
	url := fmt.Sprintf("%s:%d", Config.Indices.Host, Config.Indices.Port)
	client, err := influx.NewUDPClient(influx.UDPConfig{Addr: url})
	IndexDB = client
	return err
}
