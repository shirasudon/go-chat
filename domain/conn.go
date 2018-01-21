package domain

import "github.com/shirasudon/go-chat/domain/event"

//go:generate mockgen -destination=../internal/mocks/mock_conn.go -package=mocks github.com/shirasudon/go-chat/domain Conn

// Conn is a interface for the end-point connection for
// sending domain event.
type Conn interface {
	// It returns user specific id to distinguish which client
	// connect to.
	UserID() uint64

	// It sends any domain event to client.
	Send(ev event.Event)

	// Close close the underlying connection.
	// It should not panic when it is called multiple time, returnning error is OK.
	Close() error
}
