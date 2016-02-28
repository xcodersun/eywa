package models

import (
	"fmt"
	. "github.com/vivowares/eywa/Godeps/_workspace/src/gopkg.in/olivere/elastic.v3"
	. "github.com/vivowares/eywa/configs"
	. "github.com/vivowares/eywa/loggers"
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
	if err != nil && !Config().Indices.Disable {
		return err
	}
	_, _, err = client.Ping(url).Do()
	if err != nil && !Config().Indices.Disable {
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
