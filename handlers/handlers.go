package handlers

import (
	"github.com/vivowares/eywa/Godeps/_workspace/src/github.com/gorilla/websocket"
	"github.com/vivowares/eywa/Godeps/_workspace/src/github.com/zenazn/goji/web"
	. "github.com/vivowares/eywa/connections"
	. "github.com/vivowares/eywa/message_handlers"
	"github.com/vivowares/eywa/models"
)

var upgrader = &websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 4096,
}

func findCachedChannel(c web.C, idName string) (*models.Channel, bool) {
	id := models.DecodeHashId(c.URLParams[idName])
	ch, found := models.FetchCachedChannelById(id)
	return ch, found
}

func findChannel(c web.C) (*models.Channel, bool) {
	id := models.DecodeHashId(c.URLParams["id"])
	ch := &models.Channel{}
	found := ch.FindById(id)

	return ch, found
}

func messageHandler(ch *models.Channel) MessageHandler {
	md := NewMiddlewareStack()
	for _, hStr := range ch.MessageHandlers {
		if m, found := SupportedMessageHandlers[hStr]; found {
			md.Use(m)
		}
	}
	return md.Chain(nil)
}
