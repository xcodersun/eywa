package models

import (
	"fmt"
	. "github.com/vivowares/octopus/Godeps/_workspace/src/gopkg.in/olivere/elastic.v3"
	. "github.com/vivowares/octopus/configs"
)

var IndexClient *Client

func CloseIndexClient() error {
	return nil
}

func InitializeIndexClient() error {
	url := fmt.Sprintf("http://%s:%d", Config().Indices.Host, Config().Indices.Port)
	client, err := NewClient(
		SetURL(url),
	)
	if err != nil {
		return err
	}
	_, _, err = client.Ping(url).Do()
	if err != nil {
		return err
	}
	IndexClient = client
	return nil
}
