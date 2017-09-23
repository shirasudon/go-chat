package model

// Conn is a interface for the end-point connection for
// sending messages.
// One Conn corresponds to one connection for browser-side client.
type Conn interface {
	// it returns user specific id to distinguish users.
	UserID() uint64
	// send any ActionMessage to client.
	Send(ActionMessage)
}
