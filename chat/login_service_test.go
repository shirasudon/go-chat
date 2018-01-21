package chat

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"

	"github.com/shirasudon/go-chat/chat/queried"
	"github.com/shirasudon/go-chat/internal/mocks"
)

func TestLoginServiceImplement(t *testing.T) {
	t.Parallel()
	// make sure the interface is implemented.
	var _ LoginService = &LoginServiceImpl{}
}

func TestNewLoginServicePanic(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	testPanic := func(doFunc func()) {
		t.Helper()
		defer func() {
			if rec := recover(); rec == nil {
				t.Errorf("passing nil argument but no panic")
			}
		}()
		doFunc()
	}

	ps := mocks.NewMockPubsub(ctrl)
	users := mocks.NewMockUserQueryer(ctrl)
	testPanic(func() { _ = NewLoginServiceImpl(users, nil) })
	testPanic(func() { _ = NewLoginServiceImpl(nil, ps) })
}

func TestLoginServiceLogin(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ps := mocks.NewMockPubsub(ctrl)
	ps.EXPECT().Pub(gomock.Any()).Times(1)

	const (
		UserName = "name"
		Password = "password"
	)
	auth := queried.AuthUser{ID: 1, Name: UserName, Password: Password}

	users := mocks.NewMockUserQueryer(ctrl)
	users.EXPECT().FindByNameAndPassword(gomock.Any(), UserName, Password).
		Return(&auth, nil).Times(1)

	impl := NewLoginServiceImpl(users, ps)
	got, err := impl.Login(context.Background(), UserName, Password)
	if err != nil {
		t.Fatal(err)
	}
	if (*got) != auth {
		t.Errorf("different AuthUser, expect: %v, got: %v", got, auth)
	}
}

func TestLoginServiceLoginFail(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	const (
		UserName = "name"
		Password = "password"
	)

	ps := mocks.NewMockPubsub(ctrl)
	users := mocks.NewMockUserQueryer(ctrl)
	users.EXPECT().FindByNameAndPassword(gomock.Any(), UserName, Password).
		Return(nil, NewNotFoundError("error!")).Times(1)

	impl := NewLoginServiceImpl(users, ps)
	_, err := impl.Login(context.Background(), UserName, Password)
	if err == nil {
		t.Errorf("user not found but no error")
	}
}

func TestLoginServiceLogout(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ps := mocks.NewMockPubsub(ctrl)
	ps.EXPECT().Pub(gomock.Any()).Times(1)

	users := mocks.NewMockUserQueryer(ctrl)

	impl := NewLoginServiceImpl(users, ps)
	const (
		UserID uint64 = 1
	)
	impl.Logout(context.Background(), UserID)
}
