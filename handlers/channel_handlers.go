package handlers

import (
	"encoding/json"
	"github.com/vivowares/eywa/Godeps/_workspace/src/github.com/zenazn/goji/web"
	"github.com/vivowares/eywa/models"
	. "github.com/vivowares/eywa/presenters"
	. "github.com/vivowares/eywa/utils"
	"net/http"
)

func CreateChannel(c web.C, w http.ResponseWriter, r *http.Request) {
	ch := &models.Channel{}
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
		Render.JSON(w, http.StatusCreated, NewChannelBrief(ch))
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
	chs := models.Channels()

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

	q := &models.StatsQuery{Channel: ch}
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
	stats, found := models.FetchCachedChannelIndexStatsById(ch.Id)
	if !found {
		Render.JSON(w, http.StatusNotFound, map[string]string{"error": "channel stats not found"})
	} else {
		Render.JSON(w, http.StatusOK, stats)
	}
}

func GetChannelRequestTemplate(c web.C, w http.ResponseWriter, r *http.Request) {
	ch, found := findChannel(c)
	if !found {
		Render.JSON(w, http.StatusNotFound, map[string]string{"error": "channel not found"})
		return
	}

	query := QueryToMap(r.URL.Query())

	if len(query) != 0 {
		Render.JSON(w, http.StatusBadRequest, map[string]string{"error": "invalid query"})
		return
	}

	_, tmpl, err := FetchRequestTemplateByChannel(ch)
	if err != nil {
		Render.JSON(w, http.StatusInternalServerError, err.Error())
	} else {
		Render.Text(w, http.StatusOK, tmpl)
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
			if r.URL.Query().Get("with_indices") == "true" {
				err = ch.DeleteIndices()
			}
			if err != nil {
				Render.JSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			} else {
				w.WriteHeader(http.StatusOK)
			}
		}
	}
}
