package models

import (
	"fmt"
	influx "github.com/influxdb/influxdb/client/v2"
	. "github.com/vivowares/octopus/configs"
	"net/url"
)

var IndexDB influx.Client

func CloseIndexDB() error {
	return IndexDB.Close()
}

func InitializeIndexDB() error {
	host, err := url.Parse(
		fmt.Sprintf("http://%s:%d",
			Config.Indices.Host,
			Config.Indices.Port,
		),
	)
	if err != nil {
		return err
	}

	conf := influx.Config{URL: host}
	IndexDB = influx.NewClient(conf)

	return nil
}
