package handlers

import (
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/spf13/viper"
	"github.com/vivowares/octopus/connections"
	. "github.com/vivowares/octopus/utils"
	"net/http"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  viper.GetInt("connections.buffer_sizes.read"),
	WriteBufferSize: viper.GetInt("connections.buffer_sizes.write"),
}

func WsHandler(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		Render.JSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	vars := mux.Vars(r)
	deviceId := vars["device_id"]
	_, err = connections.CM.NewConnection(deviceId, ws, nil)
	if err != nil {
		Render.JSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
}
