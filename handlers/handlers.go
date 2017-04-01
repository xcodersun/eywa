package handlers

import (
	"github.com/gorilla/websocket"
	"github.com/zenazn/goji/web"
	. "github.com/eywa/connections"
	. "github.com/eywa/message_handlers"
	"github.com/eywa/models"
	"net/http"
)

var upgrader = &websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 4096,
	//TODO(alex): Workaround for JS to connect. Need to improve for security.
	CheckOrigin: func(r *http.Request) bool { return true },
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
	set := make(map[string]struct{})
	for _, hStr := range DefaultMessageHandlers {
		set[hStr] = struct{}{}
		md.Use(SupportedMessageHandlers[hStr])
	}

	for _, hStr := range ch.MessageHandlers {
		if _, found := set[hStr]; !found {
			if h, ok := SupportedMessageHandlers[hStr]; ok {
				set[hStr] = struct{}{}
				md.Use(h)
			}
		}
	}

	return md.Chain(nil)
}
