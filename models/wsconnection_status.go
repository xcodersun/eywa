package models

import (
	"encoding/json"
	"github.com/vivowares/eywa/Godeps/_workspace/src/gopkg.in/olivere/elastic.v3"
	"github.com/vivowares/eywa/connections"
	. "github.com/vivowares/eywa/utils"
	"time"
)

var HistoryLength = 100

type WebSocketConnectionStatus struct {
	ChannelName  string                        `json:"channel"`
	Status       string                        `json:"status"`
	ConnectedAt  *time.Time                    `json:"connected_at,omitempty"`
	LastPingedAt *time.Time                    `json:"last_pinged_at,omitempty"`
	Identifier   string                        `json:"identifier"`
	Metadata     map[string]interface{}        `json:"metadata,omitempty"`
	Histories    []*WebSocketConnectionHistory `json:"histories,omitempty"`
}

type WebSocketConnectionHistory struct {
	Activity  string    `json:"activity"`
	Timestamp time.Time `json:"timestamp"`
}

func FindWebSocketConnectionStatus(ch *Channel, devId string, withHistory bool) *WebSocketConnectionStatus {
	s := &WebSocketConnectionStatus{
		ChannelName: ch.Name,
		Status:      "offline",
		Identifier:  devId,
		Histories:   make([]*WebSocketConnectionHistory, 0),
	}

	conn, found := connections.FindWeSocketConnection(devId)
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
		orQs = append(orQs, elastic.NewTermQuery("message_type", "start"))
		orQs = append(orQs, elastic.NewTermQuery("message_type", "close"))
		boolQ.Should(orQs...)

		resp, err := IndexClient.Search().Index(GlobIndexName(ch)).Type(IndexType).Query(boolQ).Sort("timestamp", false).Size(HistoryLength).Do()
		if err == nil {
			for _, hit := range resp.Hits.Hits {
				var t map[string]interface{}
				err = json.Unmarshal(*hit.Source, &t)
				if err == nil {
					s.Histories = append(s.Histories, &WebSocketConnectionHistory{
						Activity: t["message_type"].(string),
						Timestamp: time.Unix(
							MilliSecToSec(int64(t["timestamp"].(float64))),
							MilliSecToNano(int64(t["timestamp"].(float64))),
						).UTC(),
					})
				}
			}
		}
	}

	return s
}
