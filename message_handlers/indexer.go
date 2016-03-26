package message_handlers

import (
	"fmt"
	"github.com/vivowares/eywa/Godeps/_workspace/src/github.com/satori/go.uuid"
	. "github.com/vivowares/eywa/configs"
	. "github.com/vivowares/eywa/connections"
	. "github.com/vivowares/eywa/loggers"
	. "github.com/vivowares/eywa/models"
)

var SupportedMessageHandlers = map[string]*Middleware{"indexer": Indexer}

var Indexer = NewMiddleware("indexer", func(h MessageHandler) MessageHandler {
	fn := func(c Connection, m Message, e error) {
		if !Config().Indices.Disable && e == nil && m != nil {
			if ch, found := findCachedChannel(c.ConnectionManager().Id()); found {
				id := uuid.NewV1().String()
				p, err := NewPoint(id, ch, c, m)
				if err == nil {
					p.Metadata(c.Metadata())
					_, err := IndexClient.Index().
						Index(TimedIndexName(ch, p.Timestamp)).
						Type(p.IndexType()).
						Id(id).
						BodyJson(p).
						Do()
					if err != nil {
						Logger.Error(fmt.Sprintf("error indexing point, %s", err.Error()))
					}
				} else {
					Logger.Error(fmt.Sprintf("error creating point, %s", err.Error()))
				}
			} else {
				// TODO
			}
		}

		h(c, m, e)
	}
	return MessageHandler(fn)
})
