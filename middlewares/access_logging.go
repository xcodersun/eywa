package middlewares

import (
	"bytes"
	"fmt"
	"github.com/vivowares/eywa/Godeps/_workspace/src/github.com/zenazn/goji/web"
	"github.com/vivowares/eywa/Godeps/_workspace/src/github.com/zenazn/goji/web/middleware"
	"github.com/vivowares/eywa/Godeps/_workspace/src/github.com/zenazn/goji/web/mutil"
	. "github.com/vivowares/eywa/loggers"
	"net/http"
	"net/url"
	"time"
)

func AccessLogging(c *web.C, h http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		reqID := middleware.GetReqID(*c)

		logStart(reqID, r)

		lw := mutil.WrapWriter(w)
		buf := bytes.Buffer{}
		lw.Tee(&buf)

		t1 := time.Now()
		h.ServeHTTP(lw, r)
		t2 := time.Now()

		logEnd(reqID, lw, &buf, t2.Sub(t1))
	}

	return http.HandlerFunc(fn)
}

func logStart(reqID string, r *http.Request) {
	req, err := url.QueryUnescape(r.URL.String())
	if err != nil {
		Logger.Warn(fmt.Sprintf("[%s] Started %s %s from %s, err=%s", reqID, r.Method, r.URL.String(), r.RemoteAddr, err.Error()))
	} else {
		Logger.Info(fmt.Sprintf("[%s] Started %s %s from %s", reqID, r.Method, req, r.RemoteAddr))
	}
}

func logEnd(reqID string, w mutil.WriterProxy, tee *bytes.Buffer, dt time.Duration) {
	status := w.Status()
	if status == 0 {
		status = 200
	}

	if status < 400 {
		Logger.Info(fmt.Sprintf("[%s] Returning %03d in %s", reqID, status, dt))
		Logger.Debug(fmt.Sprintf("[%s] Returning %03d in %s, with response: %s", reqID, status, dt, tee.String()))
	} else if status >= 400 && status < 500 {
		Logger.Warn(fmt.Sprintf("[%s] Returning %03d in %s, with response: %s", reqID, status, dt, tee.String()))
	} else {
		Logger.Error(fmt.Sprintf("[%s] Returning %03d in %s, with response: %s", reqID, status, dt, tee.String()))
	}
}
