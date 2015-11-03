package connections

type MessageHandler func(Connection, *Message, error)

type Middleware func(MessageHandler) MessageHandler

type Middlewares []Middleware

func (ms Middlewares) Chain(h MessageHandler) MessageHandler {
	for i := len(ms) - 1; i >= 0; i-- {
		h = ms[i](h)
	}
	return h
}
