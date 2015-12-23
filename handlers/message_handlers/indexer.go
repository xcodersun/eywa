package handlers

import (
	. "github.com/vivowares/octopus/connections"
	// . "github.com/vivowares/octopus/models"
)

var Indexer = NewMiddleware("indexer", func(h MessageHandler) MessageHandler {
	fn := func(c *Connection, m *Message, e error) {
		if e != nil {
			// fmt.Errorf("Error: %s\n", e.Error())
		} else {
			// fmt.Printf("Info: Connection: %+v\t\tMessage: %+v\n", c, m)
		}

		h(c, m, e)
	}
	return MessageHandler(fn)
})

var SupportedMessageHandlers = map[string]*Middleware{"indexer": Indexer}
