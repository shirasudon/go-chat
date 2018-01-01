package event

import (
	"context"
	"time"
)

//go:generate mockgen -destination=../../internal/mocks/mock_events.go -package=mocks github.com/shirasudon/go-chat/domain/event EventRepository

//go:generate stringer -type Type

// EventRepository is a event data store which allows only create action.
type EventRepository interface {
	// store events to the data-store.
	// It returns stored event's IDs and error if any.
	Store(ctx context.Context, ev ...Event) ([]uint64, error)
}

// Event is a domain event which is emitted when
// domain objects, such as User, Room and Message, are
// modified.
type Event interface {
	// return its type
	Type() Type

	// return its stream ID
	StreamID() StreamID

	// return its time stamp.
	Timestamp() time.Time
}

// Type represents event type.
type Type uint

const (
	TypeNone Type = iota
	TypeErrorRaised
	TypeUserCreated
	TypeUserDeleted
	TypeUserAddedFriend
	TypeRoomCreated
	TypeRoomDeleted
	TypeRoomAddedMember
	TypeRoomRemovedMember
	TypeRoomMessagesReadByUser
	TypeMessageCreated
	TypeActiveClientActivated
	TypeActiveClientInactivated
)

// StreamID represents identification for what type of domain-entity.
type StreamID uint

const (
	NoneStream StreamID = iota
	UserStream
	RoomStream
	MessageStream
)

// Common embeded fields for Event.
// It implements Event interface.
type EventEmbd struct {
	CreatedAt time.Time `json:"created_at"`
}

// Occurs confirms the event has occured at a point.
func (e *EventEmbd) Occurs() { e.CreatedAt = time.Now() }

func (EventEmbd) Type() Type             { return TypeNone }
func (EventEmbd) StreamID() StreamID     { return NoneStream }
func (e EventEmbd) Timestamp() time.Time { return e.CreatedAt }

// domain event for the error is raised.
type ErrorRaised struct {
	EventEmbd
	Message string `json:"message"`
}

func (ErrorRaised) Type() Type { return TypeErrorRaised }
