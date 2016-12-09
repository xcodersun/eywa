package handlers

import (
	"encoding/json"
	"github.com/zenazn/goji/web"
	"github.com/eywa/configs"
	. "github.com/eywa/utils"
	"net/http"
)

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
