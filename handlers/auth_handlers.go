package handlers

import (
	"github.com/vivowares/octopus/Godeps/_workspace/src/github.com/zenazn/goji/web"
	. "github.com/vivowares/octopus/configs"
	. "github.com/vivowares/octopus/models"
	. "github.com/vivowares/octopus/utils"
	"net/http"
)

func Login(c web.C, w http.ResponseWriter, r *http.Request) {
	u, p, ok := r.BasicAuth()
	if ok {
		if validateUserPassword(u, p) {
			auth, err := NewAuthToken(u, p)
			if err != nil {
				Render.JSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			} else {
				h, err := auth.Encrypt()
				if err != nil {
					Render.JSON(w, http.StatusUnauthorized, map[string]string{"error": err.Error()})
				} else {
					Render.JSON(w, http.StatusOK, map[string]interface{}{"auth_token": h, "expires_at": NanoToMilli(auth.ExpiresAt.UTC().UnixNano())})
				}
			}
		} else {
			w.WriteHeader(http.StatusUnauthorized)
		}
	} else {
		Render.JSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid BasicAuth header"})
	}
}

func validateUserPassword(u, p string) bool {
	return u == Config().Security.Dashboard.Username && p == Config().Security.Dashboard.Password
}
