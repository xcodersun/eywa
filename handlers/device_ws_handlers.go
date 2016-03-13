package handlers

import (
	"github.com/vivowares/eywa/Godeps/_workspace/src/github.com/gorilla/websocket"
	"github.com/vivowares/eywa/Godeps/_workspace/src/github.com/zenazn/goji/web"
	. "github.com/vivowares/eywa/configs"
	"github.com/vivowares/eywa/connections"
	. "github.com/vivowares/eywa/message_handlers"
	"github.com/vivowares/eywa/models"
	. "github.com/vivowares/eywa/utils"
	"net/http"
	"strings"
)

var upgrader *websocket.Upgrader

func InitWsUpgrader() {
	upgrader = &websocket.Upgrader{
		ReadBufferSize:  Config().Connections.Websocket.BufferSizes.Read,
		WriteBufferSize: Config().Connections.Websocket.BufferSizes.Write,
	}
}

func WsHandler(c web.C, w http.ResponseWriter, r *http.Request) {
	ch, found := findCachedChannel(c, "channel_id")
	if !found {
		Render.JSON(w, http.StatusNotFound, map[string]string{"error": "channel not found"})
		return
	}

	t := r.Header.Get("AccessToken")
	if len(t) == 0 || !StringSliceContains(ch.AccessTokens, t) {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	deviceId := c.URLParams["device_id"]
	if len(deviceId) == 0 {
		Render.JSON(w, http.StatusBadRequest, map[string]string{"error": "empty device id"})
		return
	}

	md := connections.NewMiddlewareStack()
	for _, hStr := range ch.MessageHandlers {
		if m, found := SupportedMessageHandlers[hStr]; found {
			md.Use(m)
		}
	}
	h := md.Chain(nil)

	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		Render.JSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	meta := QueryToMap(r.URL.Query())
	meta["_ip"] = strings.Split(r.RemoteAddr, ":")[0]
	_, err = connections.NewWebsocketConnection(deviceId, ws, h, map[string]interface{}{
		"channel":  ch,
		"metadata": meta,
	})

	if err != nil {
		Render.JSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
}

func findCachedChannel(c web.C, idName string) (*models.Channel, bool) {
	id := models.DecodeHashId(c.URLParams[idName])
	ch, found := models.FetchCachedChannelById(id)
	return ch, found
}
