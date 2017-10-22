package model

import (
	"github.com/shirasudon/go-chat/domain"
)

//go:generate mockgen -destination=../mocks/mock_pubsub.go -package=mocks github.com/shirasudon/go-chat/model Pubsub

// interface for the publisher/subcriber pattern.
type Pubsub interface {
	Pub(...domain.Event)
	Sub(domain.EventType) chan interface{}
}
