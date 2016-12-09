package handlers

import (
	"encoding/json"
	"github.com/zenazn/goji/web"
	"github.com/eywa/models"
	. "github.com/eywa/presenters"
	. "github.com/eywa/utils"
	"net/http"
	"strconv"
)

func CreateDashboard(c web.C, w http.ResponseWriter, r *http.Request) {
	dashboard := &models.Dashboard{}
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
	id, err := strconv.Atoi(c.URLParams["id"])
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	dashboard := &models.Dashboard{}
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
	dashs := models.Dashboards()

	db := []*DashboardBrief{}
	for _, d := range dashs {
		db = append(db, NewDashboardBrief(d))
	}

	Render.JSON(w, http.StatusOK, db)
}

func GetDashboard(c web.C, w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(c.URLParams["id"])
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	dashboard := &models.Dashboard{}
	found := dashboard.FindById(id)
	if !found {
		w.WriteHeader(http.StatusNotFound)
	} else {
		Render.JSON(w, http.StatusOK, dashboard)
	}
}

func DeleteDashboard(c web.C, w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(c.URLParams["id"])
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	dashboard := &models.Dashboard{Id: id}
	err = dashboard.Delete()
	if err != nil {
		Render.JSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
	} else {
		w.WriteHeader(http.StatusOK)
	}
}
