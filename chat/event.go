package chat

import "github.com/shirasudon/go-chat/domain/event"

// These events are external new types and are used only this package.

// Event for User logged in.
type eventUserLoggedIn struct {
	event.ExternalEventEmbd
	UserID uint64 `json:"user_id"`
}

func (eventUserLoggedIn) TypeString() string { return "type_user_logged_in" }

// Event for User logged out.
type eventUserLoggedOut eventUserLoggedIn

func (eventUserLoggedOut) TypeString() string { return "type_user_logged_out" }
