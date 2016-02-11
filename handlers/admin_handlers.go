package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/vivowares/octopus/Godeps/_workspace/src/github.com/zenazn/goji/web"
	"github.com/vivowares/octopus/configs"
	"github.com/vivowares/octopus/connections"
	. "github.com/vivowares/octopus/utils"
	"net/http"
	"strconv"
)

func ConnectionCounts(c web.C, w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, strconv.Itoa(connections.WebSocketCount()))
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
