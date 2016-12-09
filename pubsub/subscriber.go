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
			if s.flush([]byte(msg)) != nil {
				return
			}
		}
	}()
}
