package message_handlers

import (
	. "github.com/vivowares/eywa/connections"
	"github.com/vivowares/eywa/pubsub"
)

var Logger = NewMiddleware("logger", func(h MessageHandler) MessageHandler {
	fn := func(c Connection, m Message, e error) {
		pub := c.(pubsub.Publisher)

		if e != nil {
			pub.Publish(formatError(e))
		}

		if m != nil {
			pub.Publish(formatMessage(m))
		}

		h(c, m, e)
	}
	return MessageHandler(fn)
})
