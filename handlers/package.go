package handlers

import (
	. "github.com/vivowares/eywa/connections"
	. "github.com/vivowares/eywa/message_handlers"
	. "github.com/vivowares/eywa/models"
	. "github.com/vivowares/eywa/utils"
)

func findCachedChannel(c web.C, idName string) (*Channel, bool) {
	id := DecodeHashId(c.URLParams[idName])
	ch, found := FetchCachedChannelById(id)
	return ch, found
}

func messageHandler(ch *Channel) MessageHandler {
	md := NewMiddlewareStack()
	for _, hStr := range ch.MessageHandlers {
		if m, found := SupportedMessageHandlers[hStr]; found {
			md.Use(m)
		}
	}
	return md.Chain(nil)
}
