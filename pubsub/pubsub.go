package pubsub

import (
	"github.com/vivowares/eywa/Godeps/_workspace/src/github.com/olebedev/emitter"
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
	Publish(string)
}
