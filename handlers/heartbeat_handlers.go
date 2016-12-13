package handlers

import (
	"fmt"
	"github.com/zenazn/goji/web"
	"net/http"
)

func HeartBeatHttp(c web.C, w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "OK")
}

func HeartBeatWs(c web.C, w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "OK")
}
