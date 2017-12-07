package pubsub

import (
	"github.com/cskr/pubsub"

	"github.com/shirasudon/go-chat/domain/event"
)

// DefaultCapacity for the Pubsub
const DefaultCapacity = 1

type PubSub struct {
	pubsub *pubsub.PubSub
}

// New Pubsub with capacity size. If capacity
// is not given, use default insteadly.
func New(capacity ...int) *PubSub {
	if len(capacity) == 0 {
		capacity = []int{DefaultCapacity}
	}
	return &PubSub{pubsub: pubsub.New(capacity[0])}
}

// subscribes specified EventType and return message channel.
func (ps *PubSub) Sub(typ ...event.Type) chan interface{} {
	tags := make([]string, 0, len(typ))
	for _, tp := range typ {
		tags = append(tags, tp.String())
	}
	return ps.pubsub.Sub(tags...)
}

// unsubscribes specified channel which is gotten by previous Sub().
func (ps *PubSub) Unsub(ch chan interface{}) {
	ps.pubsub.Unsub(ch)
}

// publish Event to corresponding subscribers.
func (ps *PubSub) Pub(events ...event.Event) {
	for _, ev := range events {
		ps.pubsub.Pub(ev, ev.Type().String())
	}
}

func (ps *PubSub) Shutdown() {
	ps.pubsub.Shutdown()
}
