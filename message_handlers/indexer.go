package message_handlers

import (
	"fmt"
	"github.com/satori/go.uuid"
	. "github.com/vivowares/octopus/connections"
	. "github.com/vivowares/octopus/models"
	"time"
)

import "github.com/kr/pretty"

var SupportedMessageHandlers = map[string]*Middleware{"indexer": Indexer}

var Indexer = NewMiddleware("indexer", func(h MessageHandler) MessageHandler {
	fn := func(c *Connection, m *Message, e error) {
		if e == nil {
			if chItr, found := c.Metadata["channel"]; found {
				ch := chItr.(*Channel)
				id := uuid.NewV1().String()
				p, err := NewPoint(id, ch, c, m)
				if err == nil {
					resp, err := IndexClient.Index().
						Index(timedIndexName(ch, p.Timestamp)).
						Type("messages").
						Id(id).
						BodyJson(p).
						Do()
					if err != nil {
						pretty.Println(err)
					} else {
						pretty.Println(resp)
					}
				} else {
					pretty.Println(err)
				}
			}
		}

		h(c, m, e)
	}
	return MessageHandler(fn)
})

func timedIndexName(ch *Channel, ts time.Time) string {
	year, week := ts.ISOWeek()
	return fmt.Sprintf("channels.%d.%d-%d", ch.Id, year, week)
}
