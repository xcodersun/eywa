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
		Render.JSON(w, http.StatusBadRequest, map[string]MarshallableErrors{"errors": {"create": err}})
		return
	}

	errs, created := ch.Insert()
	if !created {
		Render.JSON(w, http.StatusBadRequest, map[string]MarshallableErrors{"errors": errs})
	} else {
		w.WriteHeader(http.StatusCreated)
	}
}

func UpdateChannel(c web.C, w http.ResponseWriter, r *http.Request) {
	name := c.URLParams["name"]
	ch, found := models.FindChannelByName(name)
	if !found {
		w.WriteHeader(http.StatusNotFound)
	} else {
		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(ch)
		if err != nil {
			Render.JSON(w, http.StatusBadRequest, map[string]MarshallableErrors{"errors": {"update": err}})
			return
		}

		errs, updated := ch.Update()
		if !updated {
			Render.JSON(w, http.StatusBadRequest, map[string]MarshallableErrors{"errors": errs})
		} else {
			w.WriteHeader(http.StatusOK)
		}
	}
}

func ListChannels(c web.C, w http.ResponseWriter, r *http.Request) {
	chs, err := models.FindChannels()
	if err != nil {
		Render.JSON(w, http.StatusInternalServerError, map[string]MarshallableErrors{"errors": {"list": err}})
	} else {
		Render.JSON(w, http.StatusOK, map[string]interface{}{"channels": chs})
	}
}

func GetChannel(c web.C, w http.ResponseWriter, r *http.Request) {
	name := c.URLParams["name"]
	ch, found := models.FindChannelByName(name)
	if !found {
		w.WriteHeader(http.StatusNotFound)
	} else {
		Render.JSON(w, http.StatusOK, map[string]interface{}{"channel": ch})
	}
}

func DeleteChannel(c web.C, w http.ResponseWriter, r *http.Request) {
	name := c.URLParams["name"]
	ch, found := models.FindChannelByName(name)
	if !found {
		w.WriteHeader(http.StatusNotFound)
	} else {
		err := ch.Delete()
		if err != nil {
			Render.JSON(w, http.StatusInternalServerError, map[string]MarshallableErrors{"errors": {"delete": err}})
		} else {
			w.WriteHeader(http.StatusOK)
		}
	}
}
