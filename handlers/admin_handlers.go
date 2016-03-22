package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/vivowares/eywa/Godeps/_workspace/src/github.com/zenazn/goji/web"
	"github.com/vivowares/eywa/configs"
	"github.com/vivowares/eywa/connections"
	"github.com/vivowares/eywa/models"
	. "github.com/vivowares/eywa/utils"
	"net/http"
)

func ConnectionCounts(c web.C, w http.ResponseWriter, r *http.Request) {
	Render.JSON(w, http.StatusOK, connections.Counts())
}

func ConnectionCount(c web.C, w http.ResponseWriter, r *http.Request) {
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

	Render.JSON(w, http.StatusOK, map[string]int{c.URLParams["channel_id"]: cm.Count()})
}

func ConnectionStatus(c web.C, w http.ResponseWriter, r *http.Request) {
	ch, found := findCachedChannel(c, "channel_id")
	if !found {
		Render.JSON(w, http.StatusNotFound, map[string]string{"error": "channel is not found"})
		return
	}

	devId := c.URLParams["device_id"]
	history := r.URL.Query().Get("history")

	status, err := models.FindConnectionStatus(ch, devId, history == "true")
	if err != nil {
		Render.JSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
	} else {
		Render.JSON(w, http.StatusOK, status)
	}
}

func GetConfig(c web.C, w http.ResponseWriter, r *http.Request) {
	Render.JSON(w, http.StatusOK, configs.Config())
}

func UpdateConfig(c web.C, w http.ResponseWriter, r *http.Request) {
	settings := map[string]interface{}{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&settings)
	if err != nil {
		Render.JSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	if err = configs.Update(settings); err != nil {
		Render.JSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
	} else {
		Render.JSON(w, http.StatusOK, configs.Config())
	}
}
