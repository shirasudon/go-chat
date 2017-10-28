package domain

// Conn is a interface for the end-point connection for
// sending domain event.
type Conn interface {
	// It returns user specific id to distinguish which client
	// connect to.
	UserID() uint64

	// It sends any domain event to client.
	Send(ev Event)
}
