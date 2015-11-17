package models

import (
	"fmt"
	influxClient "github.com/influxdb/influxdb/client/v2"
	"github.com/spf13/viper"
	"net/url"
)

var IStore IndexStore

type IndexStore interface {
	WritePoint(*Point) error
	Close() error
}

type influxStore struct {
	client influxClient.Client
}

func (is *influxStore) WritePoint(p *Point) error {
	bp, err := influxClient.NewBatchPoints(
		influxClient.BatchPointsConfig{Database: viper.GetString("indices.database")},
	)
	if err != nil {
		return err
	}

	pt, err := influxClient.NewPoint(p.channel.Name, p.Tags, p.Fields, p.Timestamp)
	if err != nil {
		return err
	}
	bp.AddPoint(pt)
	return is.client.Write(bp)
}

func (is *influxStore) Close() error {
	return is.client.Close()
}

func InitializeIndexStore() error {
	switch viper.GetString("indices.store") {
	case "influxdb":
		return initInfluxDbClient()
	default:
		return initInfluxDbClient()
	}
}

func CloseIndexStore() error {
	return IStore.Close()
}

func initInfluxDbClient() error {
	host, err := url.Parse(
		fmt.Sprintf("http://%s:%s",
			viper.GetString("indices.host"),
			viper.GetString("indices.port"),
		),
	)
	if err != nil {
		return err
	}

	conf := influxClient.Config{URL: host}
	conn := influxClient.NewClient(conf)
	IStore = &influxStore{
		client: conn,
	}
	return nil
}
