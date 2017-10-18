package pubsub

import (
	"context"

	"github.com/cskr/pubsub"

	"github.com/shirasudon/go-chat/domain"
)

type PubSub struct {
	pubsub *pubsub.PubSub
	ctx    context.Context
}

func New(ctx context.Context, capacity int) *PubSub {
	pubsub := &PubSub{pubsub: pubsub.New(capacity), ctx: ctx}
	// TODO cancelation using context
	return pubsub
}

func (ps *PubSub) Context() context.Context { return ps.ctx }

// subscribes specified EventType and return message channel.
func (ps *PubSub) Sub(typ domain.EventType) chan interface{} {
	return ps.pubsub.Sub(typ.String())
}

// unsubscribes specified channel which is gotten by previous Sub().
func (ps *PubSub) Unsub(ch chan interface{}) {
	ps.pubsub.Unsub(ch)
}

// publish Event to corresponding subscribers.
func (ps *PubSub) Pub(events ...domain.Event) {
	for _, ev := range events {
		ps.pubsub.Pub(ev, ev.EventType().String())
	}
}

func (ps *PubSub) Shutdown() {
	ps.pubsub.Shutdown()
}
