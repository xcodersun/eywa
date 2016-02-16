package handlers

import (
	"github.com/vivowares/octopus/Godeps/_workspace/src/github.com/zenazn/goji/web"
	. "github.com/vivowares/octopus/utils"
	. "github.com/vivowares/octopus/connections"
	. "github.com/vivowares/octopus/message_handlers"
	"net/http"
	"strconv"
	"time"
	"io/ioutil"
)

func HttpHandler(c web.C, w http.ResponseWriter, r *http.Request) {
	ch, found := findCachedChannel(c, "channel_id")

	if !found {
		Render.JSON(w, http.StatusNotFound, map[string]string{"error": "channel not found"})
		return
	}

	token := r.Header.Get("AccessToken")
	if len(token) == 0 || !StringSliceContains(ch.AccessTokens, token) {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	deviceId := c.URLParams["device_id"]
	if len(deviceId) == 0 {
		Render.JSON(w, http.StatusBadRequest, map[string]string{"error": "empty device id"})
		return
	}

	md := NewMiddlewareStack()
	for _, hStr := range ch.MessageHandlers {
		if m, found := SupportedMessageHandlers[hStr]; found {
			md.Use(m)
		}
	}
	h := md.Chain(nil)

	httpConn, err := NewHttpConnection(deviceId, h, map[string]interface{}{
		"channel":  ch,
		"metadata": QueryToMap(r.URL.Query()),
	})

	if err != nil {
		Render.JSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	payload, err := ioutil.ReadAll(r.Body)
	if err != nil {
		Render.JSON(w, http.StatusBadRequest, map[string]string{"error": "invalid payload"})
		return
	}

	msgId := strconv.FormatInt(time.Now().UnixNano(), 16)
	msg := &Message {
		MessageType: AsyncRequestMessage,
		MessageId: msgId,
		Payload: payload,
	}

	httpConn.MessageHandler()(httpConn, msg, nil)

}
