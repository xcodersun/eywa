package models

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/vivowares/eywa/Godeps/_workspace/src/gopkg.in/olivere/elastic.v3"
	"github.com/vivowares/eywa/connections"
	. "github.com/vivowares/eywa/utils"
	"time"
)

var HistoryLength = 100

type ConnectionStatus struct {
	ChannelName  string               `json:"channel"`
	Status       string               `json:"status"`
	ConnectedAt  *time.Time           `json:"connected_at,omitempty"`
	LastPingedAt *time.Time           `json:"last_pinged_at,omitempty"`
	Identifier   string               `json:"identifier"`
	Metadata     map[string]string    `json:"metadata,omitempty"`
	Histories    []*ConnectionHistory `json:"histories,omitempty"`
}

type ConnectionHistory struct {
	Activity       string    `json:"activity"`
	Timestamp      time.Time `json:"timestamp"`
	ConnectionType string    `json:"connection_type"`
	Duration       *int64    `json:"duration,omitempty"`
}

func FindConnectionStatus(ch *Channel, devId string, withHistory bool) (*ConnectionStatus, error) {
	name, err := ch.HashId()
	if err != nil {
		return nil, err
	}

	cm, found := connections.FindConnectionManager(name)
	if !found {
		return nil, errors.New(fmt.Sprintf("connection manager is not initialized for channel: %s", name))
	}

	s := &ConnectionStatus{
		ChannelName: ch.Name,
		Status:      "offline",
		Identifier:  devId,
		Histories:   make([]*ConnectionHistory, 0),
	}

	conn, found := cm.FindConnection(devId)
	if found {
		s.Status = "online"
		ct := conn.CreatedAt().UTC()
		s.ConnectedAt = &ct
		pt := conn.LastPingedAt().UTC()
		s.LastPingedAt = &pt
		s.Metadata = conn.Metadata()
	}

	if withHistory {
		boolQ := elastic.NewBoolQuery()
		mustQs := make([]elastic.Query, 0)
		mustQs = append(mustQs, elastic.NewTermQuery("device_id", devId))
		boolQ.Must(mustQs...)

		orQs := make([]elastic.Query, 0)
		orQs = append(orQs, elastic.NewTermQuery("message_type", connections.SupportedWebsocketMessageTypes[connections.TypeConnectMessage]))
		orQs = append(orQs, elastic.NewTermQuery("message_type", connections.SupportedWebsocketMessageTypes[connections.TypeDisconnectMessage]))
		boolQ.Should(orQs...)

		resp, err := IndexClient.Search().Index(GlobalIndexName(ch)).Type(IndexTypeActivities).Query(boolQ).Sort("timestamp", false).Size(HistoryLength).Do()
		if err == nil {
			for _, hit := range resp.Hits.Hits {
				var t map[string]interface{}
				err = json.Unmarshal(*hit.Source, &t)
				if err == nil {
					var d *int64
					if _f, ok := t["duration"].(float64); ok {
						_d := int64(_f)
						d = &_d
					}
					s.Histories = append(s.Histories, &ConnectionHistory{
						Activity: t["message_type"].(string),
						Timestamp: time.Unix(
							MilliSecToSec(int64(t["timestamp"].(float64))),
							MilliSecToNano(int64(t["timestamp"].(float64))),
						).UTC(),
						ConnectionType: t["connection_type"].(string),
						Duration:       d,
					})
				}
			}
		}
	}

	return s, nil
}
