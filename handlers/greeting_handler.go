package handlers

import (
	"github.com/zenazn/goji/web"
	. "github.com/eywa/utils"
	"net/http"
)

func Greeting(c web.C, w http.ResponseWriter, r *http.Request) {
	Render.JSON(w, http.StatusOK, map[string]string{"greeting": "I See You..."})
}
