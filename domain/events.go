package domain

import (
	"github.com/shirasudon/go-chat/domain/event"
)

// EventHolder holds event objects.
// It is used to embed into entity.
type EventHolder struct {
	events []event.Event
}

func NewEventHolder() EventHolder {
	return EventHolder{
		events: make([]event.Event, 0, 2),
	}
}

func (holder *EventHolder) Events() []event.Event {
	if holder.events == nil {
		holder.events = make([]event.Event, 0, 2)
	}
	newEvents := make([]event.Event, 0, len(holder.events))
	for _, ev := range holder.events {
		newEvents = append(newEvents, ev)
	}
	return newEvents
}

func (holder *EventHolder) AddEvent(ev event.Event) {
	if holder.events == nil {
		holder.events = make([]event.Event, 0, 2)
	}
	holder.events = append(holder.events, ev)
}
