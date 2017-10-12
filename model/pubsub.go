package model

import "github.com/shirasudon/go-chat/domain"

// interface for the publisher/subcriber pattern.
type Pubsub interface {
	Pub(...domain.Event)
	Sub(domain.EventType) chan interface{}
}
