package connections

import (
// "fmt"
)

var DefaultMiddlewares = &Middlewares{
	middlewares: []*Middleware{Logger},
}

var Logger = &Middleware{
	name: "logger",
	middleware: func(h MessageHandler) MessageHandler {
		fn := func(c *Connection, m *Message, e error) {
			if e != nil {
				// fmt.Errorf("Error: %s\n", e.Error())
			} else {
				// fmt.Printf("Info: Connection: %+v\t\tMessage: %+v\n", c, m)
			}

			h(c, m, e)
		}
		return MessageHandler(fn)
	},
}
