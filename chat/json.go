package chat

import (
	"github.com/shirasudon/go-chat/domain"
)

// EventJSON is a data-transfer-object
// which represents domain event to sent to the client connection.
// It implement Event interface.
type EventJSON struct {
	EventName string      `json:"event"`
	Data      interface{} `json:"data"`
}

const (
	EventNameMessageCreated = "message_created"
	EventNameUnknown        = "unknown"
)

var eventEncodeNames = map[domain.EventType]string{
	domain.EventMessageCreated: EventNameMessageCreated,
}

func (EventJSON) EventType() domain.EventType { return domain.EventNone }

func NewEventJSON(ev domain.Event) EventJSON {
	if ev == nil {
		panic("nil Event is not allowed")
	}

	eventName, ok := eventEncodeNames[ev.EventType()]
	if !ok {
		eventName = EventNameUnknown
	}
	return EventJSON{
		EventName: eventName,
		Data:      ev,
	}
}
