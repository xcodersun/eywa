package handlers

import (
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
	err := q.Parse(queryToMap(map[string][]string(r.URL.Query())))
	if err != nil {
		Render.JSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
	} else {
		value, err := q.QueryES()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			Render.JSON(w, http.StatusOK, map[string]interface{}{"value": value})
		}
	}
}

func queryToMap(q map[string][]string) map[string]string {
	r := make(map[string]string)
	for k, v := range q {
		if len(v) > 0 {
			r[k] = v[0]
		}
	}
	return r
}
