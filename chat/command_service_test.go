package chat

import (
	"context"
	"testing"
	"time"

	"github.com/golang/mock/gomock"

	"github.com/shirasudon/go-chat/chat/action"
	"github.com/shirasudon/go-chat/domain"
	"github.com/shirasudon/go-chat/domain/event"
	"github.com/shirasudon/go-chat/internal/mocks"
)

func TestChatUpdateServiceAtRoomDeleted(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	pubsub := mocks.NewMockPubsub(mockCtrl)

	roomDeleted := make(chan interface{}, 1)
	pubsub.EXPECT().
		Sub(event.TypeRoomDeleted).
		Return(roomDeleted).
		Times(1)

	// deleting target for the room.
	const DeletedRoomID = uint64(1)
	roomDeleteEvent := event.RoomDeleted{RoomID: DeletedRoomID}
	pubsub.EXPECT().
		Pub(roomDeleteEvent).
		Do(func(ev event.Event) {
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

func TestCommandServiceCreateRoom(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	var (
		CreateRoom = action.CreateRoom{
			SenderID:      1,
			RoomName:      "room-name",
			RoomMemberIDs: []uint64{1},
		}

		RoomID = uint64(1)

		User = domain.User{ID: CreateRoom.SenderID}
	)

	pubsub := mocks.NewMockPubsub(mockCtrl)
	pubsub.EXPECT().
		Pub(gomock.Any()).
		Times(1)

	rooms := mocks.NewMockRoomRepository(mockCtrl)
	rooms.EXPECT().
		Store(gomock.Any(), gomock.Any()).
		Return(RoomID, nil).
		Times(1)

	rooms.EXPECT().
		BeginTx(gomock.Any(), nil).
		Return(domain.EmptyTxBeginner{}, nil).
		Times(1)

	users := mocks.NewMockUserRepository(mockCtrl)
	users.EXPECT().
		Find(gomock.Any(), CreateRoom.SenderID).
		Return(User, nil).
		Times(1)

	events := mocks.NewMockEventRepository(mockCtrl)
	events.EXPECT().
		Store(gomock.Any(), gomock.Any()).
		Return([]uint64{1}, nil).
		Times(1)

	cmdService := NewCommandService(domain.SimpleRepositories{
		UserRepository:  users,
		RoomRepository:  rooms,
		EventRepository: events,
	}, pubsub)

	roomID, err := cmdService.CreateRoom(context.Background(), CreateRoom)

	if err != nil {
		t.Fatal(err)
	}
	if roomID != RoomID {
		t.Errorf("different room id for create room, expect: %v, got: %v", RoomID, roomID)
	}
}
