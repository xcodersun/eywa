package pubsub

import (
	"github.com/emitter"
)

var capacity uint = 512
var EM = emitter.New(capacity)
var EywaLogPublisher = NewBasicPublisher("log/eywa")

func Close() {
	EM.Off("*")
}

type Publisher interface {
	Topic() string
	Attached() bool
	Attach()
	Detach()
	Publish(Callback)
	Unpublish()
}

type Callback func() string
