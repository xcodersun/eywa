package handlers

import (
	"encoding/base64"
	"github.com/vivowares/eywa/Godeps/_workspace/src/github.com/gorilla/websocket"
	"github.com/vivowares/eywa/Godeps/_workspace/src/github.com/zenazn/goji/web"
	. "github.com/vivowares/eywa/configs"
	"github.com/vivowares/eywa/connections"
	. "github.com/vivowares/eywa/message_handlers"
	. "github.com/vivowares/eywa/models"
	. "github.com/vivowares/eywa/utils"
	"io/ioutil"
	"net/http"
	"strconv"
)

var upgrader *websocket.Upgrader

func InitWsUpgrader() {
	upgrader = &websocket.Upgrader{
		ReadBufferSize:  Config().WebSocketConnections.BufferSizes.Read,
		WriteBufferSize: Config().WebSocketConnections.BufferSizes.Write,
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

	_, err = connections.NewWebSocketConnection(deviceId, ws, h, map[string]interface{}{
		"channel":  ch,
		"metadata": QueryToMap(r.URL.Query()),
	})

	if err != nil {
		Render.JSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
}

func findCachedChannel(c web.C, idName string) (*Channel, bool) {
	asBytes, err := base64.URLEncoding.DecodeString(c.URLParams[idName])
	if err != nil {
		return nil, false
	}

	id, err := strconv.Atoi(string(asBytes))
	if err != nil {
		return nil, false
	}

	ch, found := FetchCachedChannelById(id)
	return ch, found
}

func SendToDevice(c web.C, w http.ResponseWriter, r *http.Request) {
	deviceId := c.URLParams["device_id"]
	conn, found := connections.FindWeSocketConnection(deviceId)
	if !found {
		Render.JSON(w, http.StatusNotFound, map[string]string{"error": "device is not online"})
		return
	}

	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		Render.JSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	err = conn.Send(bodyBytes)
	if err != nil {
		Render.JSON(w, 422, map[string]string{"error": err.Error()})
	}
}

func RequestToDevice(c web.C, w http.ResponseWriter, r *http.Request) {
	deviceId := c.URLParams["device_id"]
	conn, found := connections.FindWeSocketConnection(deviceId)
	if !found {
		Render.JSON(w, http.StatusNotFound, map[string]string{"error": "device is not online"})
		return
	}

	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		Render.JSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	msg, err := conn.Request(bodyBytes)
	if err != nil {
		Render.JSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	w.Write(msg)
}
