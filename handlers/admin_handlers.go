package handlers

import (
	"fmt"
	"github.com/vivowares/octopus/Godeps/_workspace/src/github.com/zenazn/goji/web"
	. "github.com/vivowares/octopus/connections"
	"net/http"
	"strconv"
)

func ConnectionCounts(c web.C, w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, strconv.Itoa(CM.Count()))
}
