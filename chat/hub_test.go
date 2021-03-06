package chat

import (
	"context"
	"testing"
	"time"

	"github.com/golang/mock/gomock"

	"github.com/shirasudon/go-chat/domain"
	"github.com/shirasudon/go-chat/domain/event"
	"github.com/shirasudon/go-chat/internal/mocks"
)

func TestHubImplement(t *testing.T) {
	// just check implementing interface at build time.
	var h Hub = &HubImpl{}
	_ = h
}

// SendRecorder records that the event is sent by
// using Send() method.
// It implements Conn interface.
type SendRecorder struct {
	IsSent   bool
	IsClosed bool
	userID   uint64
}

func (s *SendRecorder) UserID() uint64      { return s.userID }
func (s *SendRecorder) Send(ev event.Event) { s.IsSent = true }
func (s *SendRecorder) Close() error        { s.IsClosed = true; return nil }

func TestHubSendEvent(t *testing.T) {
	t.Parallel()

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	var (
		RoomID        = uint64(1)
		UserID        = uint64(1)
		RoomMemberIDs = []uint64{1, 2, 3}
		UserFriendIDs = []uint64{4, 5, 6}
	)

	// Firstly build Hub

	// declare mocks which returns always same object
	// except that ID is same as the argument.
	rooms := mocks.NewMockRoomRepository(mockCtrl)
	rooms.EXPECT().Find(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, roomID uint64) (domain.Room, error) {
			return domain.Room{ID: roomID, MemberIDSet: domain.NewUserIDSet(RoomMemberIDs...)}, nil
		}).AnyTimes()

	users := mocks.NewMockUserRepository(mockCtrl)
	users.EXPECT().Find(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, userID uint64) (domain.User, error) {
			return domain.User{ID: userID, FriendIDs: domain.NewUserIDSet(UserFriendIDs...)}, nil
		}).AnyTimes()

	repos := domain.SimpleRepositories{
		RoomRepository: rooms,
		UserRepository: users,
	}
	pubsub := mocks.NewMockPubsub(mockCtrl)
	pubsub.EXPECT().Pub(gomock.Any()).AnyTimes()
	commandService := NewCommandServiceImpl(repos, pubsub)

	hub := NewHubImpl(commandService)

	for _, testcase := range []struct {
		Event       event.Event
		SendUserIDs []uint64
	}{
		{
			Event:       event.MessageCreated{RoomID: RoomID},
			SendUserIDs: RoomMemberIDs,
		},
		{
			Event:       event.RoomCreated{RoomID: RoomID, MemberIDs: RoomMemberIDs},
			SendUserIDs: RoomMemberIDs,
		},
		{
			Event:       event.RoomDeleted{RoomID: RoomID, MemberIDs: RoomMemberIDs},
			SendUserIDs: RoomMemberIDs,
		},
		{
			Event:       event.RoomMessagesReadByUser{RoomID: RoomID},
			SendUserIDs: RoomMemberIDs,
		},
		{
			Event:       event.ActiveClientActivated{UserID: UserID},
			SendUserIDs: append([]uint64{1}, UserFriendIDs...), // contains UserID itself
		},
		{
			Event:       event.ActiveClientInactivated{UserID: UserID},
			SendUserIDs: UserFriendIDs,
		},
	} {
		// register user connections to Hub.
		conns := make([]*SendRecorder, 0, len(testcase.SendUserIDs))
		for _, id := range testcase.SendUserIDs {
			conn := &SendRecorder{userID: id}
			if err := hub.Connect(context.Background(), conn); err != nil {
				t.Fatalf("can not connect user id=%d, err=%v", id, err)
			}
			conns = append(conns, conn)
		}

		// send event to hub and underlying user connections..
		if err := hub.sendEvent(context.Background(), testcase.Event); err != nil {
			t.Fatalf("sending event %#v, got error: %v", testcase.Event, err)
		}

		// check every connection is sent event.
		for _, c := range conns {
			if !c.IsSent {
				t.Errorf("send %T, but user (id=%d) does not received", testcase.Event, c.UserID())
			}
			// unregister user connection which is no longer used here.
			hub.Disconnect(c)
		}
	}
}

func TestHubSendEventFail(t *testing.T) {
	t.Parallel()

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	var (
		notFoundError = NewNotFoundError("not found")
		RoomID        = uint64(1)
		UserID        = uint64(1)
		RoomMemberIDs = []uint64{1, 2, 3}
		UserFriendIDs = []uint64{4, 5, 6}
	)

	// Firstly build Hub

	// declare mocks which returns always error.
	rooms := mocks.NewMockRoomRepository(mockCtrl)
	rooms.EXPECT().Find(gomock.Any(), gomock.Any()).Return(domain.Room{}, notFoundError).AnyTimes()

	users := mocks.NewMockUserRepository(mockCtrl)
	users.EXPECT().Find(gomock.Any(), gomock.Any()).Return(domain.User{}, notFoundError).AnyTimes()

	repos := domain.SimpleRepositories{
		RoomRepository: rooms,
		UserRepository: users,
	}
	// only inactivate event is allowed.
	pubsub := mocks.NewMockPubsub(mockCtrl)
	pubsub.EXPECT().Pub(IsEvType(event.ActiveClientInactivated{})).AnyTimes()
	commandService := NewCommandServiceImpl(repos, pubsub)

	hub := NewHubImpl(commandService)

	for _, testcase := range []struct {
		Event       event.Event
		SendUserIDs []uint64
	}{
		// These events access to the repositories and gots error.
		// The other events, not accessing to the repositories, are not shown.
		{
			Event:       event.MessageCreated{RoomID: RoomID},
			SendUserIDs: RoomMemberIDs,
		},
		{
			Event:       event.RoomMessagesReadByUser{RoomID: RoomID},
			SendUserIDs: RoomMemberIDs,
		},
		{
			Event:       event.ActiveClientActivated{UserID: UserID},
			SendUserIDs: append([]uint64{1}, UserFriendIDs...), // contains UserID itself
		},
		{
			Event:       event.ActiveClientInactivated{UserID: UserID},
			SendUserIDs: UserFriendIDs,
		},
	} {
		// register user connections to Hub.
		conns := make([]*SendRecorder, 0, len(testcase.SendUserIDs))
		for _, id := range testcase.SendUserIDs {
			conn := &SendRecorder{userID: id}
			// Connect() can not be used due to mock always returns error,
			// instead of that, use Store() directly.
			ac, _, err := domain.NewActiveClient(hub.activeClients, conn, domain.User{ID: conn.UserID()})
			if err != nil {
				t.Fatalf("can not create activeClient with user id %d", conn.UserID())
			}
			if err := hub.activeClients.Store(ac); err != nil {
				t.Fatalf("can not connect user id=%d, err=%v", conn.UserID(), err)
			}
			conns = append(conns, conn)
		}

		// send event to hub and underlying user connections..
		if err := hub.sendEvent(context.Background(), testcase.Event); err == nil || err != notFoundError {
			t.Fatalf("sending event %#v, got no error or invalid error type, expect %T, got: %#v", testcase.Event, notFoundError, err)
		}

		// check every connection is sent event.
		for _, c := range conns {
			if c.IsSent {
				t.Errorf("sending %T ends error, but user (id=%d) received event", testcase.Event, c.UserID())
			}
			// unregister user connection which is no longer used here.
			hub.Disconnect(c)
		}
	}
}

func TestHubEventSendingServiceAtMessageCreated(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	pubsub := mocks.NewMockPubsub(mockCtrl)

	events := make(chan interface{}, 1)
	pubsub.EXPECT().
		Sub(HubHandlingEventTypes).
		Return(events).
		Times(1)

	const (
		UserID     = uint64(2)
		MsgContent = "hello!"
	)
	testEv := event.MessageCreated{Content: MsgContent}
	pubsub.EXPECT().
		Pub(gomock.Any()).
		AnyTimes().
		Do(func(ev event.Event) {
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
		Do(func(ev event.Event) {
			enc, ok := ev.(EventJSON)
			if !ok {
				t.Fatalf("invalid data is sent: %#v", ev)
			}

			switch enc.EventName {
			case EventNameMessageCreated:
				created, ok := enc.Data.(event.MessageCreated)
				if !ok {
					t.Fatalf("invalid data structure: %#v", enc)
				}
				if created.Content != MsgContent {
					t.Errorf("diffrent message content, expect: %v, got: %v", MsgContent, created.Content)
				}
				doneCh <- struct{}{}

			case EventNameActiveClientActivated:
				activated, ok := enc.Data.(event.ActiveClientActivated)
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
	commandService := NewCommandServiceImpl(repos, pubsub)

	// set timeout 10ms for testing.
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	hub := NewHubImpl(commandService)
	go hub.eventSendingService(ctx)
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

func TestHubActionReceivingService(t *testing.T) {
	t.Parallel()

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	const (
		UserID              = uint64(2)
		TimeoutDuration     = 10 * time.Millisecond
		ResultCheckDuration = 1 * time.Millisecond
	)

	pubsubCh := make(chan interface{}, 1)
	ps := mocks.NewMockPubsub(mockCtrl)

	ps.EXPECT().Pub(gomock.Any()).Do(func(ev event.Event) {
		pubsubCh <- ev
	}).AnyTimes()

	ps.EXPECT().Sub(event.TypeExternal).Return(pubsubCh).Times(1)

	// build mock repositories.
	users := mocks.NewMockUserRepository(mockCtrl)
	users.EXPECT().
		Find(gomock.Any(), gomock.Any()).
		Return(domain.User{ID: UserID}, nil).
		AnyTimes()

	repos := domain.SimpleRepositories{
		UserRepository: users,
	}

	// build mock conn
	conn := mocks.NewMockConn(mockCtrl)
	conn.EXPECT().
		UserID().
		Return(UserID).
		AnyTimes()
	conn.EXPECT().
		Close().
		Return(nil).
		Times(1)

	// build hub
	ctx, cancel := context.WithTimeout(context.Background(), TimeoutDuration)
	defer cancel()

	hub := NewHubImpl(NewCommandServiceImpl(repos, ps))
	go hub.actionReceivingService(ctx)

	// connect conn to Hub
	if err := hub.Connect(ctx, conn); err != nil {
		t.Fatal(err)
	}

	// pass the logout event to actionReceivingService.
	ps.Pub(eventUserLoggedOut{UserID: UserID})

	ticker := time.Tick(ResultCheckDuration)

	for {
		select {
		case <-ctx.Done():
			t.Error("timeout: logout event is occured but corresponding Conn is not removed.")
			return
		case <-ticker:
			if !hub.activeClients.ExistByConn(conn) {
				return // PASS
			}
		}
	}
}

func TestHubListenReturnByShutdown(t *testing.T) {
	t.Parallel()

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	const (
		TimeoutDuration = 10 * time.Millisecond
		SleepDuration   = 5 * time.Millisecond
	)

	pubsubCh := make(chan interface{}, 1)
	ps := mocks.NewMockPubsub(mockCtrl)
	ps.EXPECT().Sub(event.TypeExternal).Return(pubsubCh).Times(1)
	ps.EXPECT().Sub(HubHandlingEventTypes).Return(pubsubCh).Times(1)

	// build mock repositories.
	repos := domain.SimpleRepositories{}

	// build hub
	ctx, cancel := context.WithTimeout(context.Background(), TimeoutDuration)
	defer cancel()

	doneCh := make(chan bool, 1)

	hub := NewHubImpl(NewCommandServiceImpl(repos, ps))
	go func() {
		hub.Listen(ctx)
		doneCh <- true
	}()
	time.Sleep(SleepDuration) // waiting for the standing the hub.Listen

	hub.Shutdown()

	select {
	case <-ctx.Done():
		t.Error("timeout: hub shotdowned but not ends")
	case <-doneCh:
		return // PASS
	}
}
