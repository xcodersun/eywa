package pubsub

import (
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/emitter"
	"time"
)

var writeTimeout = 5 * time.Second

func NewWebsocketSubscriber(p Publisher, ws *websocket.Conn) *WebsocketSubscriber {
	return &WebsocketSubscriber{
		p:  p,
		ws: ws,
	}
}

type WebsocketSubscriber struct {
	p  Publisher
	ws *websocket.Conn
}

func (s *WebsocketSubscriber) Topic() string {
	return s.p.Topic()
}

func (s *WebsocketSubscriber) flush(p []byte) error {
	err := s.ws.SetWriteDeadline(time.Now().Add(writeTimeout))
	if err != nil {
		return err
	}
	err = s.ws.WriteMessage(websocket.TextMessage, p)
	if err != nil {
		return err
	}
	return nil
}

func (s *WebsocketSubscriber) Subscribe(banner string) {
	go func() {
		s.p.Attach()
		defer s.p.Detach()

		banner = fmt.Sprintf("%s\n%s\n\n", cowsay, banner)
		if s.flush([]byte(banner)) != nil {
			return
		}

		listener := EM.On(s.Topic(), emitter.Close)
		defer func() {
			EM.Off(s.Topic(), listener)
		}()

		for event := range listener {
			msg := event.String(0)
			if msg == "UNSUBSCRIBE" || s.flush([]byte(msg)) != nil {
				s.ws.Close();
				return
			}
		}
	}()

	// Monitor attacher's websocket
	go func() {
		messageType, _, err := s.ws.ReadMessage()
		// In error case, we want to close the websocket too. In other hand,
		// JS websocket library close with code 1005 (RFC 6455 section 11.7)
		// which is an error per gorilla websocket.
		if err != nil || messageType == websocket.CloseMessage {
			s.p.Publish(func() string{ return "UNSUBSCRIBE" })
			return
		}
	}()
}
