package model

import (
	"context"

	"github.com/shirasudon/go-chat/domain"
)

// interface for the publisher/subcriber pattern.
type Pubsub interface {
	Pub(...domain.Event)
	Sub(domain.EventType) chan interface{}

	// context for the pumping event from publisher to subcriber.
	Context() context.Context
}
