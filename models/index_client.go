package models

import (
	"fmt"
	. "gopkg.in/olivere/elastic.v3"
	. "github.com/eywa/configs"
	. "github.com/eywa/loggers"
	"log"
	"strings"
)

var IndexClient *Client

func CloseIndexClient() error {
	return nil
}

func InitializeIndexClient() error {
	url := fmt.Sprintf("http://%s:%d", Config().Indices.Host, Config().Indices.Port)
	client, err := NewClient(
		SetURL(url),
		setLogger(ESLogger),
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

func setLogger(logger *log.Logger) func(*Client) error {
	switch strings.ToUpper(Config().Logging.Indices.Level) {
	case "INFO":
		return func(c *Client) error {
			SetInfoLog(logger)
			SetErrorLog(logger)
			return nil
		}
	case "DEBUG":
		return func(c *Client) error {
			SetInfoLog(logger)
			SetErrorLog(logger)
			SetTraceLog(logger)
			return nil
		}
	default:
		return func(c *Client) error {
			SetErrorLog(logger)
			return nil
		}
	}
}
