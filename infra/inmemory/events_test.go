package inmemory

import (
	"context"
	"testing"
	"time"

	"github.com/shirasudon/go-chat/domain/event"
)

var (
	eventRepo = EventRepository{}
)

func TestEventStore(t *testing.T) {
	// case 1: store single event
	ev := event.UserCreated{}
	ev.Occurs()
	ids, err := eventRepo.Store(context.Background(), ev)
	if err != nil {
		t.Fatal(err)
	}
	if len(ids) != 1 {
		t.Fatalf("different event id size, expect: %v, got: %v", 1, len(ids))
	}

	if ids[0] != 1 {
		t.Errorf("different inserted event id, expect: %v, got: %v", 1, len(ids))
	}

	// case 2: store multiple events
	ev2 := event.RoomCreated{}
	ev2.Occurs()
	ev3 := event.MessageCreated{}
	ev3.Occurs()
	ids, err = eventRepo.Store(context.Background(), ev2, ev3)
	if err != nil {
		t.Fatal(err)
	}
	if len(ids) != 2 {
		t.Fatalf("different event id size, expect: %v, got: %v", 2, len(ids))
	}

	if ids[0] != 2 {
		t.Errorf("different inserted event id, expect: %v, got: %v", 2, len(ids))
	}
	if ids[1] != 3 {
		t.Errorf("different inserted event id, expect: %v, got: %v", 3, len(ids))
	}
}

func TestEventFindAllByTimeCursor(t *testing.T) {
	firstEvent := eventStore[0].(event.UserCreated)

	// case 1: find single result
	evs, err := eventRepo.FindAllByTimeCursor(
		context.Background(),
		firstEvent.Timestamp().Add(-1*time.Second),
		1,
	)
	if err != nil {
		t.Fatal(err)
	}

	if len(evs) != 1 {
		t.Fatalf("different returned event size, expect: %v, got: %v", 1, len(evs))
	}
	if got, ok := evs[0].(event.UserCreated); !ok || got.CreatedAt != firstEvent.CreatedAt {
		// NOTE: equalty of the event depends on Timestamp.
		t.Errorf("unexpected event is returned, expect: %v, got: %v", firstEvent, got)
	}

	// case 2: find multiple results
	secondEvent := eventStore[1].(event.RoomCreated)
	evs, err = eventRepo.FindAllByTimeCursor(context.Background(), firstEvent.Timestamp(), 2)
	if err != nil {
		t.Fatal(err)
	}

	if len(evs) != 2 {
		t.Fatalf("different returned event size, expect: %v, got: %v", 2, len(evs))
	}

	if got, ok := evs[0].(event.UserCreated); !ok || got.CreatedAt != firstEvent.CreatedAt {
		t.Errorf("unexpected event is returned, expect: %v, got: %v", firstEvent, got)
	}
	if got, ok := evs[1].(event.RoomCreated); !ok || got.CreatedAt != secondEvent.CreatedAt {
		t.Errorf("unexpected event is returned, expect: %v, got: %v", secondEvent, got)
	}
}

func TestEventFindAllByStreamID(t *testing.T) {
	firstEvent := eventStore[0].(event.UserCreated)

	// create three events to obtain at least one event for each stream.
	{
		uc := event.UserCreated{}
		uc.Occurs()
		rc := event.RoomCreated{}
		rc.Occurs()
		mc := event.MessageCreated{}
		mc.Occurs()
		_, err := eventRepo.Store(context.Background(), uc, rc, mc)
		if err != nil {
			t.Fatalf("preparation is failed: %v", err)
		}
	}

	// check the event returns correct stream id.
	for _, streamID := range []event.StreamID{
		event.UserStream,
		event.RoomStream,
		event.MessageStream,
	} {
		evs, err := eventRepo.FindAllByStreamID(
			context.Background(),
			streamID,
			firstEvent.Timestamp().Add(-1*time.Second),
			1,
		)
		if err != nil {
			t.Fatal(err)
		}

		if len(evs) != 1 {
			t.Fatalf("different returned event size, expect: %v, got: %v", 1, len(evs))
		}
		if got := evs[0].StreamID(); got != streamID {
			t.Errorf("unexpected event streamID, expect: %v, got: %v", streamID, got)
		}
	}
}
