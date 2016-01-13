package middlewares

import (
	// "bytes"
	"fmt"
	"github.com/vivowares/octopus/Godeps/_workspace/src/github.com/zenazn/goji/web"
	"github.com/vivowares/octopus/Godeps/_workspace/src/github.com/zenazn/goji/web/middleware"
	"github.com/vivowares/octopus/Godeps/_workspace/src/github.com/zenazn/goji/web/mutil"
	. "github.com/vivowares/octopus/utils"
	"net/http"
	"net/url"
	// "time"
)

func AccessLogging(c *web.C, h http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		reqID := middleware.GetReqID(*c)

		printStart(reqID, r)

		lw := mutil.WrapWriter(w)

		// t1 := time.Now()
		h.ServeHTTP(lw, r)

		if lw.Status() == 0 {
			lw.WriteHeader(http.StatusOK)
		}
		// t2 := time.Now()

		// printEnd(reqID, lw, t2.Sub(t1))
	}

	return http.HandlerFunc(fn)
}

func printStart(reqID string, r *http.Request) {
	req, err := url.QueryUnescape(r.URL.String())
	if err != nil {
		Logger.Warn(fmt.Sprintf("[%s] Started %s %s from %s, err=%s", reqID, r.Method, r.URL.String(), r.RemoteAddr, err.Error()))
	} else {
		Logger.Info(fmt.Sprintf("[%s] Started %s %s from %s", reqID, r.Method, req, r.RemoteAddr))
	}
}

// func printEnd(reqID string, w mutil.WriterProxy, dt time.Duration) {
// 	var buf bytes.Buffer

// 	cW(&buf, bBlack, "[%s] ", reqID)
// 	buf.WriteString("Returning ")
// 	status := w.Status()

// 	cW(&buf, bBlue, "%03d", status)
// 	buf.WriteString(" in ")
// 	cW(&buf, nGreen, "%s", dt)

// 	log.Print(buf.String())
// }
