package message_handlers

import (
	"fmt"
	"github.com/vivowares/eywa/Godeps/_workspace/src/github.com/satori/go.uuid"
	. "github.com/vivowares/eywa/configs"
	. "github.com/vivowares/eywa/connections"
	. "github.com/vivowares/eywa/models"
	. "github.com/vivowares/eywa/utils"
)

var SupportedMessageHandlers = map[string]*Middleware{"indexer": Indexer}

var Indexer = NewMiddleware("indexer", func(h MessageHandler) MessageHandler {
	fn := func(c Connection, m *Message, e error) {
		if !Config().Indices.Disable && e == nil {
			if chItr, found := c.Metadata()["channel"]; found {
				ch := chItr.(*Channel)
				id := uuid.NewV1().String()
				p, err := NewPoint(id, ch, c, m)
				if err == nil {
					if meta, found := c.Metadata()["metadata"]; found && meta != nil {
						p.Metadata(meta.(map[string]string))
					}

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
			}
		}

		h(c, m, e)
	}
	return MessageHandler(fn)
})
