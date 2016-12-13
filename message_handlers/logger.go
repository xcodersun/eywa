package message_handlers

import (
	. "github.com/eywa/connections"
	"github.com/eywa/pubsub"
)

var Logger = NewMiddleware("logger", func(h MessageHandler) MessageHandler {
	fn := func(c Connection, m Message, e error) {
		pub := c.(pubsub.Publisher)

		if e != nil {
			pub.Publish(func() string {
				return format("error", []byte(e.Error()))
			})
		}

		if m != nil {
			pub.Publish(func() string {
				t := m.TypeString()
				if len(t) == 0 {
					t = "wrong type"
				}

				raw := []byte{}
				if m.Raw() != nil {
					raw = m.Raw()
				}

				return format(t, raw)
			})
		}

		h(c, m, e)
	}
	return MessageHandler(fn)
})
