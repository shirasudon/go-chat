package chat

import (
	"context"
	"testing"
	"time"

	"github.com/golang/mock/gomock"

	"github.com/shirasudon/go-chat/domain"
	"github.com/shirasudon/go-chat/internal/mocks"
)

func TestHubEventSendingServiceAtMessageCreated(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	pubsub := mocks.NewMockPubsub(mockCtrl)

	events := make(chan interface{}, 1)
	pubsub.EXPECT().
		Sub(gomock.Any()).
		Return(events).
		Times(1)

	const (
		UserID     = uint64(2)
		MsgContent = "hello!"
	)
	testEv := domain.MessageCreated{Content: MsgContent}
	pubsub.EXPECT().
		Pub(gomock.Any()).
		AnyTimes().
		Do(func(ev domain.Event) {
			t.Logf("publish event: %#v", ev)
			events <- ev
		})

	// build mock repositories.
	foundRoom := domain.Room{
		MemberIDSet: domain.NewUserIDSet(UserID),
	}
	rooms := mocks.NewMockRoomRepository(mockCtrl)
	rooms.EXPECT().
		Find(gomock.Any(), gomock.Any()).
		Return(foundRoom, nil).
		Times(1)

	users := mocks.NewMockUserRepository(mockCtrl)
	users.EXPECT().
		Find(gomock.Any(), gomock.Any()).
		Return(domain.User{ID: UserID}, nil).
		AnyTimes()

	repos := domain.SimpleRepositories{
		UserRepository: users,
		RoomRepository: rooms,
	}

	// build mock conn
	conn := mocks.NewMockConn(mockCtrl)
	conn.EXPECT().
		UserID().
		Return(UserID).
		AnyTimes()

	doneCh := make(chan struct{}, 1)
	activateCh := make(chan struct{}, 1)

	conn.EXPECT().
		Send(gomock.Any()).
		Times(2).
		Do(func(ev domain.Event) {
			enc, ok := ev.(EventJSON)
			if !ok {
				t.Fatalf("invalid data is sent: %#v", ev)
			}

			switch enc.EventName {
			case EventNameMessageCreated:
				created, ok := enc.Data.(domain.MessageCreated)
				if !ok {
					t.Fatalf("invalid data structure: %#v", enc)
				}
				if created.Content != MsgContent {
					t.Errorf("diffrent message content, expect: %v, got: %v", MsgContent, created.Content)
				}
				doneCh <- struct{}{}

			case EventNameActiveClientActivated:
				activated, ok := enc.Data.(domain.ActiveClientActivated)
				if !ok {
					t.Fatalf("invalid data structure: %#v", enc)
				}
				if activated.UserID != UserID {
					t.Errorf("diffrent user id, expect: %v, got: %v", UserID, activated.UserID)
				}
				activateCh <- struct{}{}

			default:
				t.Fatalf("unknown event name, got event: %#v", enc)
			}
		})

	// build Hub
	commandService := NewCommandService(repos, pubsub)
	queryService := NewQueryService(&Queryers{
		UserQueryer: repos.Users(),
		RoomQueryer: repos.Rooms(),
	})

	// set timeout 10ms for testing.
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	hub := NewHub(commandService, queryService)
	go hub.Listen(ctx)
	defer hub.Shutdown()

	// connect conn to Hub
	if err := hub.Connect(ctx, conn); err != nil {
		t.Fatal(err)
	}

	// pass the test event to eventSendingSerive
	pubsub.Pub(testEv)

	select {
	case <-activateCh:
		select {
		case <-doneCh:
			// PASS
		case <-ctx.Done():
			t.Fatal("timeout: activated event reached but message created not reached")
		}
	case <-ctx.Done():
		t.Fatal("timeout: activated event not reached")
	}
}
