//go:generate stringer -type EventType

package domain

import "context"

//go:generate mockgen -destination=../internal/mocks/mock_events.go -package=mocks github.com/shirasudon/go-chat/domain EventRepository

type EventRepository interface {
	// store events to the data-store.
	// It returns stored event's IDs and error if any.
	Store(ctx context.Context, ev ...Event) ([]uint64, error)
}

type Event interface {
	EventType() EventType
}

// domain event for the error is raised.
type ErrorRaised struct {
	Message string `json:"message"`
}

func (ErrorRaised) EventType() EventType { return EventErrorRaised }

type EventType uint

const (
	EventNone        EventType = iota
	EventErrorRaised EventType = iota
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
