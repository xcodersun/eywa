package handlers

import (
	"github.com/vivowares/eywa/Godeps/_workspace/src/github.com/zenazn/goji/web"
	. "github.com/vivowares/eywa/configs"
	. "github.com/vivowares/eywa/connections"
	. "github.com/vivowares/eywa/message_handlers"
	. "github.com/vivowares/eywa/utils"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func HttpPushHandler(c web.C, w http.ResponseWriter, r *http.Request) {
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

	meta := QueryToMap(r.URL.Query())
	meta["_ip"] = strings.Split(r.RemoteAddr, ":")[0]
	httpConn, err := NewHttpConnection(deviceId, nil, h, map[string]interface{}{
		"channel":  ch,
		"metadata": meta,
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

	if len(payload) > 0 {
		msgId := strconv.FormatInt(time.Now().UnixNano(), 16)
		msg := &Message{
			MessageType: TypeSendMessage,
			MessageId:   msgId,
			Payload:     payload,
		}

		httpConn.MessageHandler()(httpConn, msg, nil)
	}
}

func HttpLongPollingHandler(c web.C, w http.ResponseWriter, r *http.Request) {
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

	pollCh := make(chan []byte, 1)
	meta := QueryToMap(r.URL.Query())
	meta["_ip"] = strings.Split(r.RemoteAddr, ":")[0]
	httpConn, err := NewHttpConnection(deviceId, pollCh, h, map[string]interface{}{
		"channel":  ch,
		"metadata": meta,
	})

	if err != nil {
		Render.JSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	defer httpConn.Close()

	payload, err := ioutil.ReadAll(r.Body)
	if err != nil {
		Render.JSON(w, http.StatusBadRequest, map[string]string{"error": "invalid payload"})
		return
	}

	if len(payload) > 0 {
		msgId := strconv.FormatInt(time.Now().UnixNano(), 16)
		msg := &Message{
			MessageType: TypeSendMessage,
			MessageId:   msgId,
			Payload:     payload,
		}

		httpConn.MessageHandler()(httpConn, msg, nil)
	}

	select {
	case <-HttpCloseChan:
		w.WriteHeader(http.StatusNoContent)
	case <-time.After(Config().Connections.Http.Timeouts.LongPolling.Duration):
		w.WriteHeader(http.StatusNoContent)
	case p, ok := <-pollCh:
		if ok {
			w.Write(p)
		} else {
			w.WriteHeader(http.StatusNoContent)
		}
	}
}
