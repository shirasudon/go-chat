package chat

import (
	"reflect"
	"testing"

	"github.com/shirasudon/go-chat/domain/event"
)

func TestEventEncodeNames(t *testing.T) {
	for _, type_ := range HubHandlingEventTypes {
		if _, ok := eventEncodeNames[type_]; !ok {
			t.Errorf("%v is not contained in the eventEncodeNames", type_.String())
		}
	}
}

func TestNewEventJSON(t *testing.T) {
	for _, ev := range []event.Event{
		event.MessageCreated{},
		event.ActiveClientActivated{},
		event.ActiveClientInactivated{},
		event.RoomCreated{},
		event.RoomDeleted{},
		event.RoomMessagesReadByUser{},
	} {
		evJSON := NewEventJSON(ev)
		if evJSON.EventName == EventNameUnknown {
			t.Errorf("event encode name is undefined for %T", ev)
		}
		// use reflect.DeepEqual because the event may have no comparable fields, slice or map.
		if !reflect.DeepEqual(evJSON.Data, ev) {
			t.Errorf("EventJSON has different event data for %T", ev)
		}
	}
}
