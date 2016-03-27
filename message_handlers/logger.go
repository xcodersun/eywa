package message_handlers

import (
	. "github.com/vivowares/eywa/connections"
	"github.com/vivowares/eywa/pubsub"
)

var Logger = NewMiddleware("logger", func(h MessageHandler) MessageHandler {
	fn := func(c Connection, m Message, e error) {
		pub := c.(pubsub.Publisher)

		if e != nil {
			pub.Publish(pubsub.FormatError(e))
		}

		if m != nil && m.Raw() != nil && len(m.Raw()) > 0 {
			pub.Publish(pubsub.FormatRaw(m.Raw()))
		}

		h(c, m, e)
	}
	return MessageHandler(fn)
})
