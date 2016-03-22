package handlers

import (
	"fmt"
	"github.com/vivowares/eywa/Godeps/_workspace/src/github.com/zenazn/goji/web"
	. "github.com/vivowares/eywa/configs"
	"github.com/vivowares/eywa/connections"
	. "github.com/vivowares/eywa/utils"
	"io/ioutil"
	"net/http"
	"time"
)

func SendToDevice(c web.C, w http.ResponseWriter, r *http.Request) {
	_, found := findCachedChannel(c, "channel_id")
	if !found {
		Render.JSON(w, http.StatusNotFound, map[string]string{"error": "channel is not found"})
		return
	}

	cm, found := connections.FindConnectionManager(c.URLParams["channel_id"])
	if !found {
		Render.JSON(w, http.StatusInternalServerError, map[string]string{
			"error": fmt.Sprintf("connection manager is not initialized for channel: %s", c.URLParams["channel_id"]),
		})
	}

	deviceId := c.URLParams["device_id"]
	conn, found := cm.FindConnection(deviceId)
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
	_, found := findCachedChannel(c, "channel_id")
	if !found {
		Render.JSON(w, http.StatusNotFound, map[string]string{"error": "channel is not found"})
		return
	}

	cm, found := connections.FindConnectionManager(c.URLParams["channel_id"])
	if !found {
		Render.JSON(w, http.StatusInternalServerError, map[string]string{
			"error": fmt.Sprintf("connection manager is not initialized for channel: %s", c.URLParams["channel_id"]),
		})
	}

	timeout := Config().Connections.Websocket.Timeouts.Response.Duration
	var err error
	timeoutStr := r.URL.Query().Get("timeout")
	if len(timeoutStr) > 0 {
		timeout, err = time.ParseDuration(timeoutStr)
		if err != nil {
			Render.JSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
	}

	deviceId := c.URLParams["device_id"]
	conn, found := cm.FindConnection(deviceId)
	if !found {
		Render.JSON(w, http.StatusNotFound, map[string]string{"error": "device is not online"})
		return
	}

	if wsConn, ok := conn.(*connections.WebsocketConnection); ok {
		bodyBytes, err := ioutil.ReadAll(r.Body)
		if err != nil {
			Render.JSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}

		msg, err := wsConn.Request(bodyBytes, timeout)
		if err != nil {
			Render.JSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		w.Write(msg)
	} else {
		Render.JSON(w, http.StatusBadGateway, map[string]string{"error": "only websocket connections can be requested"})
	}
}
