package handlers

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/vivowares/octopus/Godeps/_workspace/src/github.com/zenazn/goji/web"
	"github.com/vivowares/octopus/configs"
	"github.com/vivowares/octopus/connections"
	. "github.com/vivowares/octopus/models"
	. "github.com/vivowares/octopus/utils"
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
	js, err := json.MarshalIndent(configs.Config(), "", "  ")
	if err != nil {
		Render.JSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
	} else {
		fmt.Fprintf(w, "%s\n", js)
	}
}

func ReloadConfig(c web.C, w http.ResponseWriter, r *http.Request) {
	err := configs.Reload()
	if err != nil {
		Render.JSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
	} else {
		w.WriteHeader(http.StatusOK)
	}
}
