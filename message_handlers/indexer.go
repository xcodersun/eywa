package message_handlers

import (
	"errors"
	"github.com/vivowares/eywa/Godeps/_workspace/src/github.com/satori/go.uuid"
	. "github.com/vivowares/eywa/configs"
	. "github.com/vivowares/eywa/connections"
	. "github.com/vivowares/eywa/models"
)

var channelNotFound = errors.New("channel not found when indexing data")

var Indexer = NewMiddleware("indexer", func(h MessageHandler) MessageHandler {
	fn := func(c Connection, m Message, e error) {

		if !Config().Indices.Disable && e == nil && m != nil && (m.Type() == TypeUploadMessage || m.Type() == TypeDisconnectMessage || m.Type() == TypeConnectMessage) {
			if ch, found := findCachedChannel(c.ConnectionManager().Id()); found {
				id := uuid.NewV1().String()
				var p *Point
				p, e = NewPoint(id, ch, c, m)
				if e == nil {
					_, e = IndexClient.Index().
						Index(TimedIndexName(ch, p.Timestamp)).
						Type(p.IndexType()).
						Id(id).
						BodyJson(p).
						Do()
				}
			} else {
				e = channelNotFound
			}
		}

		h(c, m, e)
	}
	return MessageHandler(fn)
})
