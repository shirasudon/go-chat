package pubsub

import (
	"testing"
	"time"

	"github.com/shirasudon/go-chat/domain"
)

func TestPubsub(t *testing.T) {
	pubsub := New(10)
	defer pubsub.Shutdown()
	ch := pubsub.Sub(domain.EventRoomDeleted)

	const DeletedRoomID = 1
	pubsub.Pub(domain.RoomDeleted{RoomID: DeletedRoomID})

	timeout := time.After(1 * time.Millisecond)
	select {
	case ev := <-ch:
		if _, ok := ev.(domain.RoomDeleted); !ok {
			t.Errorf("different subsclibing event, got: %#v", ev)
		}
	case <-timeout:
		t.Error("can not subsclib the RoomDeletedEvent")
	}
}
