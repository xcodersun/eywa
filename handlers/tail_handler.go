package handlers

import (
	"github.com/zenazn/goji/web"
	"github.com/eywa/pubsub"
	. "github.com/eywa/utils"
	"net/http"
)

func TailLog(c web.C, w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		Render.JSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	pubsub.NewWebsocketSubscriber(
		pubsub.EywaLogPublisher,
		ws,
	).Subscribe("You are now attached to Eywa access log...")
}
