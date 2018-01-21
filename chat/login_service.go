package chat

import (
	"context"

	"github.com/shirasudon/go-chat/chat/queried"
)

//go:generate mockgen -destination=../internal/mocks/mock_login_service.go -package=mocks github.com/shirasudon/go-chat/chat LoginService

// LoginService is the interface for the login/logut user.
type LoginService interface {

	// Login finds authenticated user profile matched with given user name and password.
	// It returns queried user profile and nil when the user is authenticated, or
	// returns nil and NotFoundError when the user is not found.
	Login(ctx context.Context, username, password string) (*queried.AuthUser, error)

	// Logout logouts User specified userID from the chat service.
	Logout(ctx context.Context, userID uint64)
}

type LoginServiceImpl struct {
	users  UserQueryer
	pubsub Pubsub
}

func NewLoginServiceImpl(users UserQueryer, pubsub Pubsub) *LoginServiceImpl {
	if users == nil || pubsub == nil {
		panic("passing nil arguments")
	}
	return &LoginServiceImpl{
		users:  users,
		pubsub: pubsub,
	}
}

func (ls *LoginServiceImpl) Login(ctx context.Context, username, password string) (*queried.AuthUser, error) {
	auth, err := ls.users.FindByNameAndPassword(ctx, username, password)
	if err != nil {
		return nil, err
	}
	ev := eventUserLoggedIn{UserID: auth.ID}
	ev.Occurs()
	ls.pubsub.Pub(ev)
	return auth, nil
}

func (ls *LoginServiceImpl) Logout(ctx context.Context, userID uint64) {
	ev := eventUserLoggedOut{UserID: userID}
	ev.Occurs()
	ls.pubsub.Pub(ev)
}
