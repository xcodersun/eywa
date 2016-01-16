package handlers

import (
	"fmt"
	"github.com/vivowares/octopus/Godeps/_workspace/src/github.com/zenazn/goji/web"
	. "github.com/vivowares/octopus/connections"
	. "github.com/vivowares/octopus/utils"
	"io/ioutil"
	"net/http"
)

func AsyncSendToDevice(c web.C, w http.ResponseWriter, r *http.Request) {
	conn, bodyString := PreProcessRequest(c, w, r)
	if bodyString == "" {
		return
	}

	err := conn.SendAsyncRequest(bodyString)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func SyncSendToDevice(c web.C, w http.ResponseWriter, r *http.Request) {
	conn, bodyString := PreProcessRequest(c, w, r)
	if bodyString == "" {
		return
	}
	msg, err := conn.SendSyncRequest(bodyString)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	Render.JSON(w, http.StatusOK, map[string]string{"device_response": msg})
}

func PreProcessRequest(c web.C, w http.ResponseWriter, r *http.Request) (*Connection, string) {
	device_id := c.URLParams["id"]
	conn, found := CM.FindConnection(device_id)
	if !found {
		w.WriteHeader(http.StatusNotFound)
		return conn, ""
	}

	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		Render.JSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return conn, ""
	}
	return conn, string(bodyBytes)
}
