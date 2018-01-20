package server

import (
	"context"
	"testing"
	"time"

	"github.com/golang/mock/gomock"

	"github.com/shirasudon/go-chat/chat"
	"github.com/shirasudon/go-chat/domain"
	"github.com/shirasudon/go-chat/internal/mocks"
)

// Check non-panic by just calling ListenAndServe for created server.
// This can not t.Parallel becasue multiple server can not listen on same port.
func TestCreateServerFromInfra(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repos := domain.SimpleRepositories{
		UserRepository:    mocks.NewMockUserRepository(ctrl),
		RoomRepository:    mocks.NewMockRoomRepository(ctrl),
		MessageRepository: mocks.NewMockMessageRepository(ctrl),
		EventRepository:   mocks.NewMockEventRepository(ctrl),
	}

	qs := &chat.Queryers{
		UserQueryer:    mocks.NewMockUserQueryer(ctrl),
		RoomQueryer:    mocks.NewMockRoomQueryer(ctrl),
		MessageQueryer: mocks.NewMockMessageQueryer(ctrl),
	}

	ps := mocks.NewMockPubsub(ctrl)
	ps.EXPECT().Sub(gomock.Any()).AnyTimes()

	server, done := CreateServerFromInfra(repos, qs, ps, nil)
	defer done()

	doneCh := make(chan bool, 1)
	timeout := time.After(20 * time.Millisecond)
	go func() {
		server.ListenAndServe()
		close(doneCh)
	}()

	time.Sleep(10 * time.Millisecond)
	server.Shutdown(context.Background())
	select {
	case <-timeout:
		t.Error("timeout for Shutdown server")
	case <-doneCh:
		// PASS
	}
}
