package handlers

import (
	"encoding/json"
	"github.com/vivowares/octopus/models"
	. "github.com/vivowares/octopus/utils"
	"github.com/zenazn/goji/web"
	"net/http"
)

func CreateChannel(c web.C, w http.ResponseWriter, r *http.Request) {
	ch := &models.Channel{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(ch)
	if err != nil {
		Render.JSON(w, http.StatusBadRequest, map[string]map[string]error{"errors": {"create": err}})
		return
	}

	errs, created := ch.Insert()
	if !created {
		Render.JSON(w, http.StatusBadRequest, map[string]map[string]error{"errors": errs})
	} else {
		w.WriteHeader(http.StatusCreated)
	}
}

func UpdateChannel(c web.C, w http.ResponseWriter, r *http.Request) {
	ch := &models.Channel{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(ch)
	if err != nil {
		Render.JSON(w, http.StatusBadRequest, map[string]map[string]error{"errors": {"update": err}})
		return
	}

	errs, updated := ch.Update()
	if !updated {
		Render.JSON(w, http.StatusBadRequest, map[string]map[string]error{"errors": errs})
	} else {
		w.WriteHeader(http.StatusOK)
	}
}

func ListChannels(c web.C, w http.ResponseWriter, r *http.Request) {
}

func DeleteChannel(c web.C, w http.ResponseWriter, r *http.Request) {
	ch := &models.Channel{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(ch)
	if err != nil {
		Render.JSON(w, http.StatusBadRequest, map[string]map[string]error{"errors": {"delete": err}})
		return
	}

	err = ch.Delete()
	if err != nil {
		Render.JSON(w, http.StatusBadRequest, map[string]map[string]error{"errors": {"delete": err}})
		return
	}

	w.WriteHeader(http.StatusOK)
}
