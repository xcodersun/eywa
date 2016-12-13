package middlewares

import (
	"github.com/zenazn/goji/web"
	. "github.com/eywa/configs"
	. "github.com/eywa/utils"
	"net/http"
)

func ApiAuthenticator(c *web.C, h http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		if len(r.Header.Get("Api-Key")) != 0 {
			apiKey := r.Header.Get("Api-Key")
			if apiKey == Config().Security.ApiKey {
				c.Env["api_key"] = apiKey
				h.ServeHTTP(w, r)
			} else {
				Render.JSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid api key"})
			}
		} else {
			Render.JSON(w, http.StatusUnauthorized, map[string]string{"error": "empty Api-Key header"})
		}
	}

	return http.HandlerFunc(fn)
}
