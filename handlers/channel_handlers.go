package handlers

import (
	"encoding/base64"
	"encoding/json"
	"github.com/vivowares/octopus/Godeps/_workspace/src/github.com/zenazn/goji/web"
	. "github.com/vivowares/octopus/models"
	. "github.com/vivowares/octopus/presenters"
	. "github.com/vivowares/octopus/utils"
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
	asBytes, err := base64.URLEncoding.DecodeString(c.URLParams["id"])
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	id, err := strconv.Atoi(string(asBytes))
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
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

	cs := []*ChannelBrief{}
	for _, ch := range chs {
		cs = append(cs, NewChannelBrief(ch))
	}

	Render.JSON(w, http.StatusOK, cs)
}

func GetChannel(c web.C, w http.ResponseWriter, r *http.Request) {
	asBytes, err := base64.URLEncoding.DecodeString(c.URLParams["id"])
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	id, err := strconv.Atoi(string(asBytes))
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	ch := &Channel{}
	found := ch.FindById(id)
	if !found {
		w.WriteHeader(http.StatusNotFound)
	} else {
		Render.JSON(w, http.StatusOK, NewChannelDetail(ch))
	}
}

func DeleteChannel(c web.C, w http.ResponseWriter, r *http.Request) {
	asBytes, err := base64.URLEncoding.DecodeString(c.URLParams["id"])
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	id, err := strconv.Atoi(string(asBytes))
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
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
