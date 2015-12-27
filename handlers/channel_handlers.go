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
	ch, found := findChannel(c)
	if !found {
		w.WriteHeader(http.StatusNotFound)
	} else {
		fields := ch.Fields
		ch.Fields = nil
		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(ch)
		if err != nil {
			Render.JSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		if ch.Fields == nil {
			ch.Fields = fields
		}
		err = ch.Update()
		if err != nil {
			Render.JSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		} else {
			w.WriteHeader(http.StatusOK)
		}
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
	ch, found := findChannel(c)

	if !found {
		w.WriteHeader(http.StatusNotFound)
	} else {
		Render.JSON(w, http.StatusOK, NewChannelDetail(ch))
	}
}

func GetChannelTagStats(c web.C, w http.ResponseWriter, r *http.Request) {
	ch, found := findCachedChannel(c, "id")
	if !found {
		Render.JSON(w, http.StatusNotFound, map[string]string{"error": "channel not found"})
		return
	}

	q := &StatsQuery{Channel: ch}
	err := q.Parse(QueryToMap(r.URL.Query()))
	if err != nil {
		Render.JSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
	} else {
		stats, err := q.QueryES()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			Render.JSON(w, http.StatusOK, stats)
		}
	}
}

func GetChannelIndexStats(c web.C, w http.ResponseWriter, r *http.Request) {
	ch, found := findCachedChannel(c, "id")
	if !found {
		Render.JSON(w, http.StatusNotFound, map[string]string{"error": "channel not found"})
		return
	}
	stats, found := FetchCachedChannelIndexStatsById(ch.Id)
	if !found {
		Render.JSON(w, http.StatusNotFound, map[string]string{"error": "channel stats not found"})
	} else {
		Render.JSON(w, http.StatusOK, stats)
	}
}

func DeleteChannel(c web.C, w http.ResponseWriter, r *http.Request) {
	ch, found := findChannel(c)
	if !found {
		w.WriteHeader(http.StatusNotFound)
	} else {
		err := ch.Delete()
		if err != nil {
			Render.JSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		} else {
			w.WriteHeader(http.StatusOK)
		}
	}
}

func findChannel(c web.C) (*Channel, bool) {
	asBytes, err := base64.URLEncoding.DecodeString(c.URLParams["id"])
	if err != nil {
		return nil, false
	}

	id, err := strconv.Atoi(string(asBytes))
	if err != nil {
		return nil, false
	}

	ch := &Channel{}
	found := ch.FindById(id)

	return ch, found
}
