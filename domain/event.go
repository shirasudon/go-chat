//go:generate stringer -type EventType

package domain

type Event interface {
	EventType() EventType
}

type EventType uint

const (
	EventNone EventType = iota
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
