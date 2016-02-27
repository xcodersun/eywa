package middlewares

import (
	"github.com/vivowares/eywa/Godeps/_workspace/src/github.com/zenazn/goji/web"
	. "github.com/vivowares/eywa/models"
	. "github.com/vivowares/eywa/utils"
	"net/http"
)

var PublicPaths = []string{"/login", "/heartbeat", "/", ""}

func Authenticator(c *web.C, h http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		if StringSliceContains(PublicPaths, r.URL.Path) {
			h.ServeHTTP(w, r)
		} else {
			if len(r.Header.Get("Authentication")) != 0 {
				tokenStr := r.Header.Get("Authentication")
				auth, err := DecryptAuthToken(tokenStr)
				if err != nil {
					Render.JSON(w, http.StatusUnauthorized, map[string]string{"error": err.Error()})
				} else {
					c.Env["auth_token"] = auth
					h.ServeHTTP(w, r)
				}
			} else {
				w.WriteHeader(http.StatusUnauthorized)
			}
		}
	}

	return http.HandlerFunc(fn)
}
