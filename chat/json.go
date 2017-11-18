package chat

import (
	"time"

	"github.com/shirasudon/go-chat/domain"
)

const (
	EventNameMessageCreated          = "message_created"
	EventNameActiveClientActivated   = "client_activated"
	EventNameActiveClientInactivated = "client_inactivated"
	EventNameUnknown                 = "unknown"
)

var eventEncodeNames = map[domain.EventType]string{
	domain.EventMessageCreated:          EventNameMessageCreated,
	domain.EventActiveClientActivated:   EventNameActiveClientActivated,
	domain.EventActiveClientInactivated: EventNameActiveClientInactivated,
}

// EventJSON is a data-transfer-object
// which represents domain event to sent to the client connection.
// It implement Event interface.
type EventJSON struct {
	EventName string      `json:"event"`
	Data      interface{} `json:"data"`
}

func (EventJSON) EventType() domain.EventType { return domain.EventNone }
func (e EventJSON) Timestamp() time.Time      { return e.Data.(domain.Event).Timestamp() }

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
