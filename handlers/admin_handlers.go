package handlers

import (
	"encoding/json"
	"github.com/vivowares/eywa/Godeps/_workspace/src/github.com/zenazn/goji/web"
	"github.com/vivowares/eywa/configs"
	"github.com/vivowares/eywa/connections"
	"github.com/vivowares/eywa/models"
	. "github.com/vivowares/eywa/utils"
	"net/http"
)

func ConnectionCounts(c web.C, w http.ResponseWriter, r *http.Request) {
	Render.JSON(w, http.StatusOK, map[string]int{"count": connections.Count()})
}

func ConnectionStatus(c web.C, w http.ResponseWriter, r *http.Request) {
	chId := c.URLParams["channel_id"]
	devId := c.URLParams["device_id"]
	history := r.URL.Query().Get("history")

	id := models.DecodeHashId(chId)
	ch := &models.Channel{}
	found := ch.FindById(id)
	if !found {
		Render.JSON(w, http.StatusNotFound, map[string]string{"error": "channel not found"})
	}

	status := models.FindConnectionStatus(ch, devId, history == "true")
	Render.JSON(w, http.StatusOK, status)
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
