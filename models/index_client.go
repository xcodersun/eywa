package models

import (
	"fmt"
	. "github.com/vivowares/octopus/configs"
	. "gopkg.in/olivere/elastic.v3"
)

var IndexClient *Client

func CloseIndexClient() error {
	return nil
}

func InitializeIndexClient() error {
	client, err := NewClient(
		SetURL(fmt.Sprintf("http://%s:%d", Config.Indices.Host, Config.Indices.Port)),
	)
	if err != nil {
		return err
	}
	// _, _, err = client.Ping().Do()
	// if err != nil {
	// 	return err
	// }
	IndexClient = client
	return nil
}
