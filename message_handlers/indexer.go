package message_handlers

import (
	"fmt"
	"github.com/vivowares/octopus/Godeps/_workspace/src/github.com/satori/go.uuid"
	. "github.com/vivowares/octopus/connections"
	. "github.com/vivowares/octopus/models"
)

var SupportedMessageHandlers = map[string]*Middleware{"indexer": Indexer}

var Indexer = NewMiddleware("indexer", func(h MessageHandler) MessageHandler {
	fn := func(c *Connection, m *Message, e error) {
		if e == nil {
			if chItr, found := c.Metadata["channel"]; found {
				ch := chItr.(*Channel)
				id := uuid.NewV1().String()
				p, err := NewPoint(id, ch, c, m)
				if err == nil {
					if meta, found := c.Metadata["metadata"]; found && meta != nil {
						p.Metadata(meta.(map[string]string))
					}

					resp, err := IndexClient.Index().
						Index(TimedIndexName(ch, p.Timestamp)).
						Type(IndexType).
						Id(id).
						BodyJson(p).
						Do()
					if err != nil {
						fmt.Println(err)
					} else {
						fmt.Println(resp)
					}
				} else {
					fmt.Println(err)
				}
			}
		}

		h(c, m, e)
	}
	return MessageHandler(fn)
})
