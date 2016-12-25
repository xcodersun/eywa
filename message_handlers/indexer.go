package message_handlers

import (
	"encoding/json"
	"github.com/satori/go.uuid"
	"gopkg.in/olivere/elastic.v3"
	. "github.com/eywa/configs"
	. "github.com/eywa/connections"
	. "github.com/eywa/models"
	"github.com/eywa/pubsub"
)

var Indexer = NewMiddleware("indexer", func(h MessageHandler) MessageHandler {
	fn := func(c Connection, m Message, e error) {
		if !Config().Indices.Disable && e == nil && m != nil && (m.Type() == TypeUploadMessage || m.Type() == TypeDisconnectMessage || m.Type() == TypeConnectMessage) {
			if ch, found := findCachedChannel(c.ConnectionManager().Id()); found {
				id := uuid.NewV1().String()
				var p *Point
				p, e = NewPoint(id, ch, c, m)
				if e == nil {
					var js []byte
					js, e = json.Marshal(p)
					if e == nil {
						var resp *elastic.IndexResponse
						resp, e = IndexClient.Index().
							Index(TimedIndexName(ch, p.Timestamp)).
							Type(p.IndexType()).
							Id(id).
							BodyString(string(js)).
							Do()

						if resp != nil && resp.Created {
							c.(pubsub.Publisher).Publish(func() string {
								return format("index", js)
							})
						}
					}
				}
			} else {
				e = channelNotFound
			}
		}

		h(c, m, e)
	}
	return MessageHandler(fn)
})
