package pubsub

import (
	"sync/atomic"
)

type BasicPublisher struct {
	attachers int32

	topic string
}

func NewBasicPublisher(topic string) *BasicPublisher {
	return &BasicPublisher{
		attachers: 0,
		topic:     topic,
	}
}

func (p *BasicPublisher) Attach() { atomic.AddInt32(&p.attachers, 1) }

func (p *BasicPublisher) Detach() { atomic.AddInt32(&p.attachers, -1) }

func (p *BasicPublisher) Attached() bool { return atomic.LoadInt32(&p.attachers) > 0 }

func (p *BasicPublisher) Topic() string { return p.topic }

func (p *BasicPublisher) Publish(c Callback) {
	if p.Attached() {
		EM.Emit(p.Topic(), c())
	}
}

func (p *BasicPublisher) Unpublish() {
	EM.Off(p.Topic())
}
