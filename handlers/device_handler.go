package handlers

import (
	"github.com/vivowares/octopus/Godeps/_workspace/src/github.com/zenazn/goji/web"
	"github.com/vivowares/octopus/connections"
	. "github.com/vivowares/octopus/utils"
	"io/ioutil"
	"net/http"
)

func AsyncSendToDevice(c web.C, w http.ResponseWriter, r *http.Request) {
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

	err = conn.SendAsyncRequest(bodyBytes)
	if err != nil {
		Render.JSON(w, 422, map[string]string{"error": err.Error()})
	}
}

func SyncSendToDevice(c web.C, w http.ResponseWriter, r *http.Request) {
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

	msg, err := conn.SendSyncRequest(bodyBytes)
	if err != nil {
		Render.JSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	w.Write(msg)
}
