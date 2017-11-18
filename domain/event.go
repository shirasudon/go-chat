//go:generate stringer -type EventType

package domain

import (
	"context"
	"time"
)

//go:generate mockgen -destination=../internal/mocks/mock_events.go -package=mocks github.com/shirasudon/go-chat/domain EventRepository

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
	EventType() EventType

	// return its time stamp.
	Timestamp() time.Time
}

type EventType uint

const (
	EventNone EventType = iota
	EventErrorRaised
	EventUserCreated
	EventUserDeleted
	EventUserAddedFriend
	EventRoomCreated
	EventRoomDeleted
	EventRoomAddedMember
	EventRoomRemoveMember
	EventRoomPostedMessage
	EventRoomUpdatedMessage
	EventRoomDeletedMessage
	EventMessageCreated
	EventMessageReadByUser
	EventActiveClientActivated
	EventActiveClientInactivated
)

// Common embeded fields for Event.
// It implements Event interface.
type EventEmbd struct {
	CreatedAt time.Time `json:"created_at"`
}

// Occurs confirms the event has occured at a point.
func (e *EventEmbd) Occurs() { e.CreatedAt = time.Now() }

func (EventEmbd) EventType() EventType   { return EventNone }
func (e EventEmbd) Timestamp() time.Time { return e.CreatedAt }

// domain event for the error is raised.
type ErrorRaised struct {
	EventEmbd
	Message string `json:"message"`
}

func (ErrorRaised) EventType() EventType { return EventErrorRaised }

// EventHolder holds event objects.
// It is used to embed into entity.
type EventHolder struct {
	events []Event
}

func NewEventHolder() EventHolder {
	return EventHolder{
		events: make([]Event, 0, 2),
	}
}

func (holder *EventHolder) Events() []Event {
	if holder.events == nil {
		holder.events = make([]Event, 0, 2)
	}
	newEvents := make([]Event, 0, len(holder.events))
	for _, ev := range holder.events {
		newEvents = append(newEvents, ev)
	}
	return newEvents
}

func (holder *EventHolder) AddEvent(ev Event) {
	if holder.events == nil {
		holder.events = make([]Event, 0, 2)
	}
	holder.events = append(holder.events, ev)
}
