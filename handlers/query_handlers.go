package handlers

import (
	"fmt"
	"github.com/vivowares/octopus/Godeps/_workspace/src/github.com/zenazn/goji/web"
	. "github.com/vivowares/octopus/models"
	. "github.com/vivowares/octopus/utils"
	"net/http"
)

func QueryValue(c web.C, w http.ResponseWriter, r *http.Request) {
	ch, found := findCachedChannel(c, "id")
	if !found {
		Render.JSON(w, http.StatusNotFound, map[string]string{"error": "channel not found"})
		return
	}

	q := &ValueQuery{Channel: ch}
	err := q.Parse(QueryToMap(r.URL.Query()))
	if err != nil {
		Render.JSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
	} else {
		value, err := q.QueryES()
		if err != nil {
			Render.Text(w, http.StatusInternalServerError, err.Error())
		} else {
			Render.JSON(w, http.StatusOK, value)
		}
	}
}

func QuerySeries(c web.C, w http.ResponseWriter, r *http.Request) {
	ch, found := findCachedChannel(c, "id")
	if !found {
		Render.JSON(w, http.StatusNotFound, map[string]string{"error": "channel not found"})
		return
	}

	q := &SeriesQuery{Channel: ch}
	err := q.Parse(QueryToMap(r.URL.Query()))
	if err != nil {
		Render.JSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
	} else {
		series, err := q.QueryES()
		if err != nil {
			Render.Text(w, http.StatusInternalServerError, err.Error())
		} else {
			Render.JSON(w, http.StatusOK, series)
		}
	}
}

func QueryRaw(c web.C, w http.ResponseWriter, r *http.Request) {
	ch, found := findCachedChannel(c, "id")
	if !found {
		Render.JSON(w, http.StatusNotFound, map[string]string{"error": "channel not found"})
		return
	}

	q := &RawQuery{Channel: ch}
	err := q.Parse(QueryToMap(r.URL.Query()))
	if err != nil {
		Render.JSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
	} else {
		bytes, err := q.QueryESNop()
		if err != nil {
			Render.Text(w, http.StatusInternalServerError, err.Error())
		} else {
			mega := bytes / 1024 / 1024
			Render.JSON(w, http.StatusOK, map[string]string{"size": fmt.Sprintf("%dmb", mega)})
		}
	}
}
