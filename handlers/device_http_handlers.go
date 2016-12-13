package handlers

import (
	"fmt"
	"github.com/zenazn/goji/web"
	"github.com/zenazn/goji/web/middleware"
	. "github.com/eywa/configs"
	. "github.com/eywa/connections"
	. "github.com/eywa/utils"
	"net/http"
	"strings"
	"time"
)

func HttpPushHandler(c web.C, w http.ResponseWriter, r *http.Request) {
	ch, found := findCachedChannel(c, "channel_id")

	if !found {
		Render.JSON(w, http.StatusNotFound, map[string]string{"error": "channel not found"})
		return
	}

	token := r.Header.Get("AccessToken")
	if len(token) == 0 || !StringSliceContains(ch.AccessTokens, token) {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	deviceId := c.URLParams["device_id"]
	if len(deviceId) == 0 {
		Render.JSON(w, http.StatusBadRequest, map[string]string{"error": "empty device id"})
		return
	}

	cm, found := FindConnectionManager(c.URLParams["channel_id"])
	if !found {
		Render.JSON(w, http.StatusNotFound, map[string]string{
			"error": fmt.Sprintf("connection manager is not initialized for channel %s", c.URLParams["channel_id"]),
		})
		return
	}

	meta := QueryToMap(r.URL.Query())
	meta["ip"] = strings.Split(r.RemoteAddr, ":")[0]
	meta["request_id"] = c.Env[middleware.RequestIDKey].(string)

	conn, err := HttpUp.Upgrade(w, r, HttpPush)
	if err != nil {
		Render.JSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	_, err = cm.NewHttpConnection(deviceId, conn, messageHandler(ch), meta)
	if err != nil {
		Render.JSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
}

func HttpLongPollingHandler(c web.C, w http.ResponseWriter, r *http.Request) {
	ch, found := findCachedChannel(c, "channel_id")

	if !found {
		Render.JSON(w, http.StatusNotFound, map[string]string{"error": "channel not found"})
		return
	}

	token := r.Header.Get("AccessToken")
	if len(token) == 0 || !StringSliceContains(ch.AccessTokens, token) {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	deviceId := c.URLParams["device_id"]
	if len(deviceId) == 0 {
		Render.JSON(w, http.StatusBadRequest, map[string]string{"error": "empty device id"})
		return
	}

	cm, found := FindConnectionManager(c.URLParams["channel_id"])
	if !found {
		Render.JSON(w, http.StatusNotFound, map[string]string{
			"error": fmt.Sprintf("connection manager is not initialized for channel: %s", c.URLParams["channel_id"]),
		})
		return
	}

	meta := QueryToMap(r.URL.Query())

	timeout := Config().Connections.Http.Timeouts.LongPolling.Duration
	if timeoutStr, found := meta["timeout"]; found {
		delete(meta, "timeout")
		var err error
		timeout, err = time.ParseDuration(timeoutStr)
		if err != nil {
			Render.JSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
	}

	meta["ip"] = strings.Split(r.RemoteAddr, ":")[0]
	meta["request_id"] = c.Env[middleware.RequestIDKey].(string)

	conn, err := HttpUp.Upgrade(w, r, HttpPoll)
	if err != nil {
		Render.JSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	httpConn, err := cm.NewHttpConnection(deviceId, conn, messageHandler(ch), meta)
	if err != nil {
		Render.JSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	resp := httpConn.Poll(timeout)

	if resp == nil {
		w.WriteHeader(http.StatusNoContent)
	} else {
		w.Write(resp)
	}
}
