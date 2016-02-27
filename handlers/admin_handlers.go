package handlers

import (
	"encoding/base64"
	"encoding/json"
	"github.com/vivowares/eywa/Godeps/_workspace/src/github.com/zenazn/goji/web"
	"github.com/vivowares/eywa/configs"
	"github.com/vivowares/eywa/connections"
	. "github.com/vivowares/eywa/models"
	"github.com/vivowares/eywa/presenters"
	. "github.com/vivowares/eywa/utils"
	"net/http"
	"strconv"
)

func ConnectionCounts(c web.C, w http.ResponseWriter, r *http.Request) {
	Render.JSON(w, http.StatusOK, map[string]int{"count": connections.WebSocketCount()})
}

func ConnectionStatus(c web.C, w http.ResponseWriter, r *http.Request) {
	chId := c.URLParams["channel_id"]
	devId := c.URLParams["device_id"]
	history := r.URL.Query().Get("history")

	asBytes, err := base64.URLEncoding.DecodeString(chId)
	if err != nil {
		Render.JSON(w, http.StatusNotFound, map[string]string{"error": "channel not found"})
	}

	id, err := strconv.Atoi(string(asBytes))
	if err != nil {
		Render.JSON(w, http.StatusNotFound, map[string]string{"error": "channel not found"})
	}

	ch := &Channel{}
	found := ch.FindById(id)
	if !found {
		Render.JSON(w, http.StatusNotFound, map[string]string{"error": "channel not found"})
	}

	status := FindWebSocketConnectionStatus(ch, devId, history == "true")
	Render.JSON(w, http.StatusOK, status)
}

func GetConfig(c web.C, w http.ResponseWriter, r *http.Request) {
	Render.JSON(w, http.StatusOK, presenters.NewConf(configs.Config()))
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
		Render.JSON(w, http.StatusOK, presenters.NewConf(configs.Config()))
	}
}
