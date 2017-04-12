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

// Message handler in fact is a stack of middlewares.
func messageHandler(ch *models.Channel) MessageHandler {
	// Initiate a new stack.
	md := NewMiddlewareStack()
	set := make(map[string]struct{})
	// Go through the default message handlers and chain them together.
	for _, hStr := range DefaultMessageHandlers {
		set[hStr] = struct{}{}
		// append the handler to the last one.
		md.Use(SupportedMessageHandlers[hStr])
	}

	// Go through the custom message handlers.
	for _, hStr := range ch.MessageHandlers {
		if _, found := set[hStr]; !found {
			if h, ok := SupportedMessageHandlers[hStr]; ok {
				set[hStr] = struct{}{}
				// append the handler to the last one.
				md.Use(h)
			}
		}
	}
	// Ends the message handler chain with empty function.
	return md.Chain(nil)
}
