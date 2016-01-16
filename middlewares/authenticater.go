package middlewares

import (
	"github.com/vivowares/octopus/Godeps/_workspace/src/github.com/zenazn/goji/web"
	. "github.com/vivowares/octopus/models"
	. "github.com/vivowares/octopus/utils"
	"net/http"
)

func Authenticator(c *web.C, h http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/login" {
			h.ServeHTTP(w, r)
		} else {
			if len(r.Header.Get("AuthToken")) != 0 {
				tokenStr := r.Header.Get("AuthToken")
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
