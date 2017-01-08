package handlers

import (
	"github.com/zenazn/goji/web"
	"github.com/eywa/models"
	. "github.com/eywa/utils"
	"github.com/eywa/connections"
	"net/http"
)

func GetSummary(c web.C, w http.ResponseWriter, r *http.Request) {
	resp := make(map[string]int);

	params := r.URL.Query()
	if len(params) == 0 {
		chs := models.Channels()
		das := models.Dashboards()
		_, total := connections.Counts()
		resp["channels"] = len(chs)
		resp["dashboards"] = len(das)
		resp["devices"] = total
	} else {
		if _, found := params["channels"]; found {
			chs := models.Channels()
			resp["channels"] = len(chs)
		} else if _, found = params["dashboards"]; found {
			das := models.Dashboards()
			resp["dashboards"] = len(das)
		} else if _, found = params["devices"]; found {
			_, total := connections.Counts()
			resp["devices"] = total
		} else {
			Render.JSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		}
	}

	Render.JSON(w, http.StatusOK, resp)
}
