package connections

import (
	"reflect"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestMiddlewares(t *testing.T) {
	Convey("stack the middlewares in the correct order", t, func() {
		m1 := &Middleware{name: "m1", handlerFunc: nil}
		m2 := &Middleware{name: "m2", handlerFunc: nil}
		m3 := &Middleware{name: "m3", handlerFunc: nil}
		m4 := &Middleware{name: "m4", handlerFunc: nil}
		m5 := &Middleware{name: "m5", handlerFunc: nil}
		m6 := &Middleware{name: "m6", handlerFunc: nil}
		m7 := &Middleware{name: "m7", handlerFunc: nil}

		ms := &MiddlewareStack{middlewares: []*Middleware{m2}}

		ms.Use(m1)
		ms.InsertBefore(m7, m1)
		ms.InsertBefore(m5, m2)
		ms.InsertAfter(m3, m1)
		ms.InsertAfter(m4, m5)
		ms.InsertAfter(m6, m2)
		ms.Remove(m2)

		names := []string{}
		for _, m := range ms.middlewares {
			names = append(names, m.name)
		}
		So(reflect.DeepEqual(names, []string{
			"m5", "m4", "m6", "m7", "m1", "m3",
		}), ShouldBeTrue)
	})

	Convey("chain the middlewares in the correct order", t, func() {
		ms := &MiddlewareStack{middlewares: []*Middleware{}}
		array := []string{}

		m1 := &Middleware{
			name: "m1",
			handlerFunc: func(h MessageHandler) MessageHandler {
				fn := func(c Connection, m Message, e error) {
					array = append(array, "<m1>")
					h(c, m, e)
					array = append(array, "</m1>")
				}
				return MessageHandler(fn)
			},
		}

		m2 := &Middleware{
			name: "m1",
			handlerFunc: func(h MessageHandler) MessageHandler {
				fn := func(c Connection, m Message, e error) {
					array = append(array, "<m2>")
					h(c, m, e)
					array = append(array, "</m2>")
				}
				return MessageHandler(fn)
			},
		}

		ms.Use(m1)
		ms.InsertBefore(m2, m1)
		h := func(c Connection, m Message, e error) {
			array = append(array, "h")
		}
		ms.Chain(MessageHandler(h))(nil, nil, nil)
		So(reflect.DeepEqual(array,
			[]string{"<m2>", "<m1>", "h", "</m1>", "</m2>"}),
			ShouldBeTrue)
	})
}
