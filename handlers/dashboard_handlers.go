package handlers

import (
	"encoding/base64"
	"encoding/json"
	"github.com/vivowares/eywa/Godeps/_workspace/src/github.com/zenazn/goji/web"
	. "github.com/vivowares/eywa/models"
	. "github.com/vivowares/eywa/presenters"
	. "github.com/vivowares/eywa/utils"
	"net/http"
	"strconv"
)

func CreateDashboard(c web.C, w http.ResponseWriter, r *http.Request) {
	dashboard := &Dashboard{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(dashboard)
	if err != nil {
		Render.JSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	err = dashboard.Create()
	if err != nil {
		Render.JSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
	} else {
		w.WriteHeader(http.StatusCreated)
	}
}

func UpdateDashboard(c web.C, w http.ResponseWriter, r *http.Request) {
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

	dashboard := &Dashboard{}
	found := dashboard.FindById(id)
	if !found {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(dashboard)
	if err != nil {
		Render.JSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	err = dashboard.Update()
	if err != nil {
		Render.JSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
	} else {
		w.WriteHeader(http.StatusOK)
	}

}

func ListDashboards(c web.C, w http.ResponseWriter, r *http.Request) {
	ds := []*Dashboard{}
	DB.Find(&ds)

	db := []*DashboardBrief{}
	for _, d := range ds {
		db = append(db, NewDashboardBrief(d))
	}

	Render.JSON(w, http.StatusOK, db)
}

func GetDashboard(c web.C, w http.ResponseWriter, r *http.Request) {
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

	dashboard := &Dashboard{}
	found := dashboard.FindById(id)
	if !found {
		w.WriteHeader(http.StatusNotFound)
	} else {
		Render.JSON(w, http.StatusOK, NewDashboardDetail(dashboard))
	}
}

func DeleteDashboard(c web.C, w http.ResponseWriter, r *http.Request) {
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

	dashboard := &Dashboard{Id: id}
	err = dashboard.Delete()
	if err != nil {
		Render.JSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
	} else {
		w.WriteHeader(http.StatusOK)
	}
}
