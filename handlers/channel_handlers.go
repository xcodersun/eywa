package handlers

import (
	"encoding/json"
	. "github.com/vivowares/octopus/models"
	. "github.com/vivowares/octopus/utils"
	"github.com/zenazn/goji/web"
	"net/http"
	"strconv"
)

func CreateChannel(c web.C, w http.ResponseWriter, r *http.Request) {
	ch := &Channel{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(ch)
	if err != nil {
		Render.JSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	err = ch.Create()
	if err != nil {
		Render.JSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
	} else {
		w.WriteHeader(http.StatusCreated)
	}
}

func UpdateChannel(c web.C, w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(c.URLParams["id"])
	if err != nil {
		Render.JSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	ch := &Channel{}
	found := ch.FindById(id)
	if !found {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(ch)
	if err != nil {
		Render.JSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	err = ch.Update()
	if err != nil {
		Render.JSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
	} else {
		w.WriteHeader(http.StatusOK)
	}

}

func ListChannels(c web.C, w http.ResponseWriter, r *http.Request) {
	chs := []*Channel{}
	DB.Find(&chs)

	Render.JSON(w, http.StatusOK, chs)
}

func GetChannel(c web.C, w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(c.URLParams["id"])
	if err != nil {
		Render.JSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	ch := &Channel{}
	found := ch.FindById(id)
	if !found {
		w.WriteHeader(http.StatusNotFound)
	} else {
		Render.JSON(w, http.StatusOK, ch)
	}
}

func DeleteChannel(c web.C, w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(c.URLParams["id"])
	if err != nil {
		Render.JSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	ch := &Channel{Id: id}
	err = ch.Delete()
	if err != nil {
		Render.JSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
	} else {
		w.WriteHeader(http.StatusOK)
	}
}
