package pubsub

import (
	"fmt"
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
	go func() {
		defer func() { recover() }()

		fmt.Printf("all subscribers: %d\n", EM.Listeners("*"))

		if p.Attached() {
			fmt.Printf("publishing %s:  %d\n", p.Topic()+"/*", len(EM.Listeners(p.Topic()+"/*")))
			EM.Emit(p.Topic()+"/*", c())
		}
	}()
}

func (p *BasicPublisher) Unpublish() {
	if p.Attached() {
		EM.Off(p.Topic() + "/*")
		fmt.Printf("unpublishing %s:  %d\n", p.Topic()+"/*", len(EM.Listeners(p.Topic()+"/*")))
	}
}
