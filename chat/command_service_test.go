package chat

import (
	"context"
	"testing"
	"time"

	"github.com/golang/mock/gomock"

	"github.com/shirasudon/go-chat/domain"
	"github.com/shirasudon/go-chat/mocks"
)

func TestChatUpdateServiceAtRoomDeleted(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	pubsub := mocks.NewMockPubsub(mockCtrl)

	roomDeleted := make(chan interface{}, 1)
	pubsub.EXPECT().
		Sub(domain.EventRoomDeleted).
		Return(roomDeleted).
		Times(1)

	// deleting target for the room.
	const DeletedRoomID = uint64(1)
	roomDeleteEvent := domain.RoomDeleted{RoomID: DeletedRoomID}
	pubsub.EXPECT().
		Pub(roomDeleteEvent).
		Do(func(ev domain.Event) {
			t.Log("Pubsub.Pub(roomDeleteEvent)")
			roomDeleted <- ev
		})

	// set timeout 10ms for testing.
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	doneCh := make(chan struct{}, 1)

	messages := mocks.NewMockMessageRepository(mockCtrl)
	messages.EXPECT().
		RemoveAllByRoomID(ctx, roomDeleteEvent.RoomID).
		Return(nil).
		Times(1).
		Do(func(ctx context.Context, roomID uint64) {
			doneCh <- struct{}{}
		})

	commandService := NewCommandService(domain.SimpleRepositories{
		MessageRepository: messages,
	}, pubsub)

	go commandService.RunUpdateService(ctx)
	defer commandService.CancelUpdateService()

	// pass the roomDeleteEvent to updateService
	pubsub.Pub(roomDeleteEvent)

	select {
	case <-doneCh:
		// PASS
	case <-ctx.Done():
		t.Errorf("timeout to fail")
	}
}
