package pubsub

import (
	"testing"
	"time"

	"github.com/shirasudon/go-chat/domain/event"
)

func TestPubsub(t *testing.T) {
	pubsub := New(10)
	defer pubsub.Shutdown()
	ch := pubsub.Sub(event.TypeRoomDeleted)

	const DeletedRoomID = 1
	pubsub.Pub(event.RoomDeleted{RoomID: DeletedRoomID})

	timeout := time.After(1 * time.Millisecond)
	select {
	case ev := <-ch:
		if _, ok := ev.(event.RoomDeleted); !ok {
			t.Errorf("different subsclibing event, got: %#v", ev)
		}
	case <-timeout:
		t.Error("can not subsclib the RoomDeletedEvent")
	}
}
