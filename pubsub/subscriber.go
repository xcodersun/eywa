package pubsub

import (
	"fmt"
	"github.com/vivowares/eywa/Godeps/_workspace/src/github.com/gorilla/websocket"
	"strings"
	"time"
)

var writeTimeout = 5 * time.Second

func NewWebsocketSubscriber(p Publisher, id string, ws *websocket.Conn) *WebsocketSubscriber {
	return &WebsocketSubscriber{
		p:  p,
		id: id,
		ws: ws,
	}
}

type WebsocketSubscriber struct {
	id string
	p  Publisher
	ws *websocket.Conn
}

func (s *WebsocketSubscriber) Topic() string {
	return s.p.Topic() + "/" + strings.Replace(s.id, "/", "-", -1)
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

		defer func() {
			s.p.Detach()
			EM.Off(s.Topic())
			fmt.Println("unsub")
		}()

		banner = fmt.Sprintf("%s\n%s\n\n", cowsay, banner)
		if s.flush([]byte(banner)) != nil {
			return
		}

		for event := range EM.On(s.Topic()) {
			msg := event.String(0)
			if s.flush([]byte(msg)) != nil {
				return
			}
		}
	}()
}
