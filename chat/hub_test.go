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

	msgCreated := make(chan interface{}, 1)
	pubsub.EXPECT().
		Sub(domain.EventMessageCreated).
		Return(msgCreated).
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
			if _, ok := ev.(domain.MessageCreated); ok {
				msgCreated <- ev
			}
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
		Times(1)

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
	conn.EXPECT().
		Send(gomock.Any()).
		Do(func(ev domain.Event) {
			enc, ok := ev.(EncodedEvent)
			if !ok {
				t.Fatalf("invalid data is sent: %#v", ev)
			}
			created, ok := enc[EncNameMessageCreated].(domain.MessageCreated)
			if !ok {
				t.Fatalf("invalid data structure: %#v", enc)
			}
			if created.Content != MsgContent {
				t.Errorf("diffrent message content, expect: %v, got: %v", MsgContent, created.Content)
			}
			doneCh <- struct{}{}
		})

	// build Hub
	commandService := NewCommandService(repos, pubsub)
	queryService := NewQueryService(repos)

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
	case <-doneCh:
		// PASS
	case <-ctx.Done():
		t.Errorf("timeout to fail")
	}
}
