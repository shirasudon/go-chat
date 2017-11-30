package chat

import (
	"time"

	"github.com/shirasudon/go-chat/domain/event"
)

const (
	EventNameMessageCreated          = "message_created"
	EventNameActiveClientActivated   = "client_activated"
	EventNameActiveClientInactivated = "client_inactivated"
	EventNameUnknown                 = "unknown"
)

var eventEncodeNames = map[event.Type]string{
	event.TypeMessageCreated:          EventNameMessageCreated,
	event.TypeActiveClientActivated:   EventNameActiveClientActivated,
	event.TypeActiveClientInactivated: EventNameActiveClientInactivated,
}

// EventJSON is a data-transfer-object
// which represents domain event to sent to the client connection.
// It implement Event interface.
type EventJSON struct {
	EventName string      `json:"event"`
	Data      event.Event `json:"data"`
}

func (EventJSON) Type() event.Type           { return event.TypeNone }
func (e EventJSON) Timestamp() time.Time     { return e.Data.Timestamp() }
func (e EventJSON) StreamID() event.StreamID { return e.Data.StreamID() }

func NewEventJSON(ev event.Event) EventJSON {
	if ev == nil {
		panic("nil Event is not allowed")
	}

	eventName, ok := eventEncodeNames[ev.Type()]
	if !ok {
		eventName = EventNameUnknown
	}
	return EventJSON{
		EventName: eventName,
		Data:      ev,
	}
}
