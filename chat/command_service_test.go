package chat

import (
	"context"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/golang/mock/gomock"

	"github.com/shirasudon/go-chat/chat/action"
	"github.com/shirasudon/go-chat/domain"
	"github.com/shirasudon/go-chat/domain/event"
	"github.com/shirasudon/go-chat/internal/mocks"
)

// EventMathcher matches type of the internal Event.
// It implements gomock.Matcher interface.
type EventMathcher struct {
	Event event.Event
}

func IsEvType(ev event.Event) EventMathcher {
	return EventMathcher{ev}
}

func (em EventMathcher) Matches(x interface{}) bool {
	return reflect.ValueOf(em.Event).Type() == reflect.ValueOf(x).Type()
}

func (em EventMathcher) String() string {
	return fmt.Sprintf("type %T", em.Event)
}

func TestCommandServiceImplement(t *testing.T) {
	// just check implementing interface at build time.
	var cs CommandService = &CommandServiceImpl{}
	_ = cs
}

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

	commandService := NewCommandServiceImpl(domain.SimpleRepositories{
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
		Pub(IsEvType(event.RoomCreated{})).
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

	cmdService := NewCommandServiceImpl(domain.SimpleRepositories{
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

func TestCommandServiceDeleteRoom(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	var (
		DeleteRoom = action.DeleteRoom{
			SenderID: 1,
			RoomID:   1,
		}

		User = domain.User{ID: DeleteRoom.SenderID}
		Room = domain.Room{ID: DeleteRoom.RoomID, OwnerID: User.ID}
	)

	pubsub := mocks.NewMockPubsub(mockCtrl)
	publishEv := pubsub.EXPECT().
		Pub(IsEvType(event.RoomDeleted{})).
		Times(1)

	rooms := mocks.NewMockRoomRepository(mockCtrl)
	roomFind := rooms.EXPECT().
		Find(gomock.Any(), DeleteRoom.RoomID).
		Return(Room, nil).
		Times(1)

	beginTx := rooms.EXPECT().
		BeginTx(gomock.Any(), gomock.Nil()).
		Return(domain.EmptyTxBeginner{}, nil).
		Times(1)

	roomRemove := rooms.EXPECT().
		Remove(gomock.Any(), Room).
		Return(nil).
		Times(1)

	gomock.InOrder(
		beginTx,
		roomFind,
		roomRemove,
		publishEv,
	)

	users := mocks.NewMockUserRepository(mockCtrl)
	users.EXPECT().
		Find(gomock.Any(), DeleteRoom.SenderID).
		Return(User, nil).
		Times(1)

	events := mocks.NewMockEventRepository(mockCtrl)
	events.EXPECT().
		Store(gomock.Any(), gomock.Any()).
		Return([]uint64{1}, nil).
		Times(1)

	cmdService := NewCommandServiceImpl(domain.SimpleRepositories{
		UserRepository:  users,
		RoomRepository:  rooms,
		EventRepository: events,
	}, pubsub)

	// do test function.
	roomID, err := cmdService.DeleteRoom(context.Background(), DeleteRoom)
	if err != nil {
		t.Fatal(err)
	}
	if roomID != Room.ID {
		t.Errorf("different room id for deleting room, expect: %v, got: %v", Room.ID, roomID)
	}
}

func TestCommandServiceAddRoomMember(t *testing.T) {
	t.Parallel()

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	var (
		AddRoomMember = action.AddRoomMember{
			SenderID:  1,
			RoomID:    1,
			AddUserID: 2,
		}

		User   = domain.User{ID: AddRoomMember.AddUserID}
		Sender = domain.User{ID: AddRoomMember.SenderID}
		Room   = domain.Room{
			ID:          AddRoomMember.RoomID,
			OwnerID:     AddRoomMember.SenderID,
			MemberIDSet: domain.NewUserIDSet(AddRoomMember.SenderID),
		}
	)

	pubsub := mocks.NewMockPubsub(mockCtrl)
	publishEv := pubsub.EXPECT().
		Pub(IsEvType(event.RoomAddedMember{})).
		Times(1)

	rooms := mocks.NewMockRoomRepository(mockCtrl)
	roomFind := rooms.EXPECT().
		Find(gomock.Any(), AddRoomMember.RoomID).
		Return(Room, nil).
		Times(1)

	beginTx := rooms.EXPECT().
		BeginTx(gomock.Any(), gomock.Nil()).
		Return(domain.EmptyTxBeginner{}, nil).
		Times(1)

	roomStore := rooms.EXPECT().
		Store(gomock.Any(), gomock.Any()). // Room is modified here.
		Return(Room.ID, nil).
		Times(1)

	gomock.InOrder(
		beginTx,
		roomFind,
		roomStore,
		publishEv,
	)

	users := mocks.NewMockUserRepository(mockCtrl)
	users.EXPECT().
		Find(gomock.Any(), AddRoomMember.SenderID).
		Return(Sender, nil).
		Times(1)

	users.EXPECT().
		Find(gomock.Any(), AddRoomMember.AddUserID).
		Return(User, nil).
		Times(1)

	events := mocks.NewMockEventRepository(mockCtrl)
	events.EXPECT().
		Store(gomock.Any(), gomock.Any()).
		Return([]uint64{1}, nil).
		Times(1)

	cmdService := NewCommandServiceImpl(domain.SimpleRepositories{
		UserRepository:  users,
		RoomRepository:  rooms,
		EventRepository: events,
	}, pubsub)

	// do test function.
	res, err := cmdService.AddRoomMember(context.Background(), AddRoomMember)
	if err != nil {
		t.Fatal(err)
	}
	if res.RoomID != Room.ID {
		t.Errorf("different room id for AddRoomMember, expect: %v, got: %v", Room.ID, res.RoomID)
	}
	if res.UserID != User.ID {
		t.Errorf("different user id for AddRoomMember, expect: %v, got: %v", User.ID, res.UserID)
	}
}

func TestCommandServicePostRoomMessage(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	var (
		ChatMessage = action.ChatMessage{
			SenderID: 1,
			RoomID:   1,
			Content:  "hello",
		}

		User = domain.User{ID: ChatMessage.SenderID}
		Room = domain.Room{ID: ChatMessage.RoomID, OwnerID: User.ID,
			MemberIDSet: domain.NewUserIDSet(User.ID)}
	)

	const (
		NewMsgID = uint64(1)
	)

	rooms := mocks.NewMockRoomRepository(mockCtrl)
	rooms.EXPECT().
		Find(gomock.Any(), ChatMessage.RoomID).
		Return(Room, nil).
		Times(1)

	users := mocks.NewMockUserRepository(mockCtrl)
	users.EXPECT().
		Find(gomock.Any(), ChatMessage.SenderID).
		Return(User, nil).
		Times(1)

	msgs := mocks.NewMockMessageRepository(mockCtrl)
	beginTx := msgs.EXPECT().
		BeginTx(gomock.Any(), gomock.Nil()).
		Return(domain.EmptyTxBeginner{}, nil).
		Times(1)

	msgStore := msgs.EXPECT().
		Store(gomock.Any(), gomock.Any()).
		Return(NewMsgID, nil).
		Times(1)

	pubsub := mocks.NewMockPubsub(mockCtrl)
	publishEv := pubsub.EXPECT().
		Pub(IsEvType(event.MessageCreated{})).
		Times(1)

	gomock.InOrder(
		beginTx,
		msgStore,
		publishEv,
	)

	events := mocks.NewMockEventRepository(mockCtrl)
	events.EXPECT().
		Store(gomock.Any(), gomock.Any()).
		Return([]uint64{1}, nil).
		Times(1)

	cmdService := NewCommandServiceImpl(domain.SimpleRepositories{
		UserRepository:    users,
		RoomRepository:    rooms,
		MessageRepository: msgs,
		EventRepository:   events,
	}, pubsub)

	// do test function.
	msgID, err := cmdService.PostRoomMessage(context.Background(), ChatMessage)
	if err != nil {
		t.Fatal(err)
	}
	if msgID != NewMsgID {
		t.Errorf("different new message id for post room message, expect: %v, got: %v", NewMsgID, msgID)
	}
}

func TestCommandServiceReadRoomMessages(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	{ // case1 success
		var (
			RoomCreated = time.Now()

			ReadMessages = action.ReadMessages{
				SenderID: 1,
				RoomID:   1,
				ReadAt:   RoomCreated.Add(24 * time.Hour),
			}

			User = domain.User{ID: ReadMessages.SenderID}
			Room = domain.Room{ID: ReadMessages.RoomID, OwnerID: User.ID, CreatedAt: RoomCreated,
				MemberIDSet: domain.NewUserIDSet(User.ID), MemberReadTimes: domain.NewTimeSet(User.ID)}
		)

		users := mocks.NewMockUserRepository(mockCtrl)
		users.EXPECT().
			Find(gomock.Any(), ReadMessages.SenderID).
			Return(User, nil).
			Times(1)

		rooms := mocks.NewMockRoomRepository(mockCtrl)
		beginTx := rooms.EXPECT().
			BeginTx(gomock.Any(), gomock.Nil()).
			Return(domain.EmptyTxBeginner{}, nil).
			Times(1)

		roomsFind := rooms.EXPECT().
			Find(gomock.Any(), ReadMessages.RoomID).
			Return(Room, nil).
			Times(1)

		roomsStore := rooms.EXPECT().
			Store(gomock.Any(), gomock.Any()).
			Return(ReadMessages.RoomID, nil).
			Times(1)

		pubsub := mocks.NewMockPubsub(mockCtrl)
		publishEv := pubsub.EXPECT().
			Pub(IsEvType(event.RoomMessagesReadByUser{})).
			Times(1)

		gomock.InOrder(
			beginTx,
			roomsFind,
			roomsStore,
			publishEv,
		)

		events := mocks.NewMockEventRepository(mockCtrl)
		events.EXPECT().
			Store(gomock.Any(), gomock.Any()).
			Return([]uint64{1}, nil).
			Times(1)

		cmdService := NewCommandServiceImpl(domain.SimpleRepositories{
			UserRepository:  users,
			RoomRepository:  rooms,
			EventRepository: events,
		}, pubsub)

		// do test function.
		roomID, err := cmdService.ReadRoomMessages(context.Background(), ReadMessages)
		if err != nil {
			t.Fatal(err)
		}
		if roomID != ReadMessages.RoomID {
			t.Errorf("different updated room id for read room message, expect: %v, got: %v", ReadMessages.RoomID, roomID)
		}
	}

	{ // case2: empty ReadAt
		var (
			RoomCreated = time.Now()

			ReadMessages = action.ReadMessages{
				SenderID: 1,
				RoomID:   1,
				ReadAt:   time.Time{},
			}

			User = domain.User{ID: ReadMessages.SenderID}
			Room = domain.Room{ID: ReadMessages.RoomID, OwnerID: User.ID, CreatedAt: RoomCreated,
				MemberIDSet: domain.NewUserIDSet(User.ID), MemberReadTimes: domain.NewTimeSet(User.ID)}
		)

		users := mocks.NewMockUserRepository(mockCtrl)
		users.EXPECT().Find(gomock.Any(), ReadMessages.SenderID).
			Return(User, nil)

		rooms := mocks.NewMockRoomRepository(mockCtrl)
		rooms.EXPECT().BeginTx(gomock.Any(), gomock.Nil()).Return(domain.EmptyTxBeginner{}, nil)
		rooms.EXPECT().Find(gomock.Any(), ReadMessages.RoomID).Return(Room, nil)
		rooms.EXPECT().Store(gomock.Any(), gomock.Any()).Return(ReadMessages.RoomID, nil)

		pubsub := mocks.NewMockPubsub(mockCtrl)
		pubsub.EXPECT().Pub(IsEvType(event.RoomMessagesReadByUser{}))

		events := mocks.NewMockEventRepository(mockCtrl)
		events.EXPECT().Store(gomock.Any(), gomock.Any()).Return([]uint64{1}, nil)

		cmdService := NewCommandServiceImpl(domain.SimpleRepositories{
			UserRepository:  users,
			RoomRepository:  rooms,
			EventRepository: events,
		}, pubsub)

		// do test function.
		roomID, err := cmdService.ReadRoomMessages(context.Background(), ReadMessages)
		if err != nil {
			t.Fatal(err)
		}
		if roomID != ReadMessages.RoomID {
			t.Errorf("different updated room id for read room message, expect: %v, got: %v", ReadMessages.RoomID, roomID)
		}
		if got, ok := Room.MemberReadTimes.Get(User.ID); !ok || got == (time.Time{}) {
			t.Errorf("after read by user, the read time in the room is not changed")
		}
	}
}
