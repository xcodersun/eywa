package handlers

import (
	"fmt"
	"github.com/zenazn/goji/web"
	"github.com/zenazn/goji/web/middleware"
	"github.com/eywa/connections"
	. "github.com/eywa/utils"
	"net/http"
	"strings"
)

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

	cm, found := connections.FindConnectionManager(c.URLParams["channel_id"])
	if !found {
		Render.JSON(w, http.StatusNotFound, map[string]string{
			"error": fmt.Sprintf("connection manager is not initialized for channel: %s", c.URLParams["channel_id"]),
		})
		return
	}

	ws, err := connections.WsUp.Upgrade(w, r, nil)
	if err != nil {
		Render.JSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	meta := QueryToMap(r.URL.Query())
	meta["ip"] = strings.Split(r.RemoteAddr, ":")[0]
	meta["request_id"] = c.Env[middleware.RequestIDKey].(string)

	_, err = cm.NewWebsocketConnection(deviceId, ws, messageHandler(ch), meta)

	if err != nil {
		Render.JSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
}
