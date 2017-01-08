package handlers

import (
	"errors"
	"fmt"
	"github.com/zenazn/goji/web"
	. "github.com/eywa/configs"
	"github.com/eywa/connections"
	"github.com/eywa/models"
	"github.com/eywa/pubsub"
	. "github.com/eywa/utils"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func ConnectionCounts(c web.C, w http.ResponseWriter, r *http.Request) {
	connectionsCounts, _ := connections.Counts()
	Render.JSON(w, http.StatusOK, connectionsCounts)
}

func ConnectionCount(c web.C, w http.ResponseWriter, r *http.Request) {
	_, found := findCachedChannel(c, "channel_id")
	if !found {
		Render.JSON(w, http.StatusNotFound, map[string]string{"error": "channel is not found"})
		return
	}

	cm, found := connections.FindConnectionManager(c.URLParams["channel_id"])
	if !found {
		Render.JSON(w, http.StatusNotFound, map[string]string{
			"error": fmt.Sprintf("connection manager is not initialized for channel: %s", c.URLParams["channel_id"]),
		})
		return
	}

	Render.JSON(w, http.StatusOK, map[string]int{c.URLParams["channel_id"]: cm.Count()})
}

func ConnectionStatus(c web.C, w http.ResponseWriter, r *http.Request) {
	ch, found := findCachedChannel(c, "channel_id")
	if !found {
		Render.JSON(w, http.StatusNotFound, map[string]string{"error": "channel is not found"})
		return
	}

	devId := c.URLParams["device_id"]
	history := r.URL.Query().Get("history")

	status, err := models.FindConnectionStatus(ch, devId, history == "true")
	if err != nil {
		Render.JSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
	} else {
		Render.JSON(w, http.StatusOK, status)
	}
}

func SendToDevice(c web.C, w http.ResponseWriter, r *http.Request) {
	_, found := findCachedChannel(c, "channel_id")
	if !found {
		Render.JSON(w, http.StatusNotFound, map[string]string{"error": "channel is not found"})
		return
	}

	cm, found := connections.FindConnectionManager(c.URLParams["channel_id"])
	if !found {
		Render.JSON(w, http.StatusNotFound, map[string]string{
			"error": fmt.Sprintf("connection manager is not initialized for channel: %s", c.URLParams["channel_id"]),
		})
		return
	}

	deviceId := c.URLParams["device_id"]
	conn, found := cm.FindConnection(deviceId)
	if !found {
		Render.JSON(w, http.StatusNotFound, map[string]string{"error": "device is not online"})
		return
	}

	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		Render.JSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	sender, ok := conn.(connections.Sender)
	if !ok {
		Render.JSON(w, http.StatusBadGateway, map[string]string{"error": errors.New("connection is not allowed to send").Error()})
		return
	}

	err = sender.Send(bodyBytes)
	if err != nil {
		Render.JSON(w, http.StatusBadGateway, map[string]string{"error": err.Error()})
	}
}

func RequestToDevice(c web.C, w http.ResponseWriter, r *http.Request) {
	_, found := findCachedChannel(c, "channel_id")
	if !found {
		Render.JSON(w, http.StatusNotFound, map[string]string{"error": "channel is not found"})
		return
	}

	cm, found := connections.FindConnectionManager(c.URLParams["channel_id"])
	if !found {
		Render.JSON(w, http.StatusNotFound, map[string]string{
			"error": fmt.Sprintf("connection manager is not initialized for channel: %s", c.URLParams["channel_id"]),
		})
		return
	}

	timeout := Config().Connections.Websocket.Timeouts.Response.Duration
	var err error
	timeoutStr := r.URL.Query().Get("timeout")
	if len(timeoutStr) > 0 {
		timeout, err = time.ParseDuration(timeoutStr)
		if err != nil {
			Render.JSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
	}

	deviceId := c.URLParams["device_id"]
	conn, found := cm.FindConnection(deviceId)
	if !found {
		Render.JSON(w, http.StatusNotFound, map[string]string{"error": "device is not online"})
		return
	}

	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		Render.JSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	requester, ok := conn.(connections.Requester)
	if !ok {
		Render.JSON(w, http.StatusBadGateway, map[string]string{"error": errors.New("connection is not allowed to request").Error()})
		return
	}

	msg, err := requester.Request(bodyBytes, timeout)
	if err != nil {
		Render.JSON(w, http.StatusBadGateway, map[string]string{"error": err.Error()})
		return
	}

	w.Write(msg)
}

func ScanConnections(c web.C, w http.ResponseWriter, r *http.Request) {
	ch, found := findCachedChannel(c, "channel_id")
	if !found {
		Render.JSON(w, http.StatusNotFound, map[string]string{"error": "channel is not found"})
		return
	}

	size := 25 //default page size
	sizeStr := r.URL.Query().Get("size")
	if len(sizeStr) > 0 {
		var err error
		size, err = strconv.Atoi(sizeStr)
		if err != nil {
			Render.JSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
	}

	lastId := strings.TrimSpace(r.URL.Query().Get("last"))

	cm, found := connections.FindConnectionManager(c.URLParams["channel_id"])
	if !found {
		Render.JSON(w, http.StatusNotFound, map[string]string{
			"error": fmt.Sprintf("connection manager is not initialized for channel: %s", c.URLParams["channel_id"]),
		})
		return
	}

	conns := cm.Scan(lastId, size)

	connSts := make([]*models.ConnectionStatus, len(conns))
	for i, conn := range conns {
		connSts[i] = models.NewConnectionStatus(ch, conn)
	}

	Render.JSON(w, http.StatusOK, connSts)
}

func AttachConnection(c web.C, w http.ResponseWriter, r *http.Request) {
	_, found := findCachedChannel(c, "channel_id")
	if !found {
		Render.JSON(w, http.StatusNotFound, map[string]string{"error": "channel is not found"})
		return
	}

	cm, found := connections.FindConnectionManager(c.URLParams["channel_id"])
	if !found {
		Render.JSON(w, http.StatusNotFound, map[string]string{
			"error": fmt.Sprintf("connection manager is not initialized for channel: %s", c.URLParams["channel_id"]),
		})
		return
	}

	deviceId := c.URLParams["device_id"]
	conn, found := cm.FindConnection(deviceId)
	if !found {
		Render.JSON(w, http.StatusNotFound, map[string]string{"error": "device is not online"})
		return
	}

	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		Render.JSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	pubsub.NewWebsocketSubscriber(
		conn.(pubsub.Publisher),
		ws,
	).Subscribe(
		fmt.Sprintf("You are now attached to connection %s:%s ...", cm.Id(), conn.Identifier()),
	)

}
