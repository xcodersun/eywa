package handlers

import (
	"fmt"
	"github.com/vivowares/octopus/Godeps/_workspace/src/github.com/zenazn/goji/web"
	"net/http"
)

func HeartBeatHttp(c web.C, w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "OK")
}

func HeartBeatWs(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "OK")
}
