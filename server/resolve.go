package server

import (
	"context"

	"github.com/shirasudon/go-chat/chat"
	"github.com/shirasudon/go-chat/domain"
)

// resolve.go provide helper functions for resolving dependencies to
// construct gochat Server.

// DoneFunc is function to be called after all of operations are done.
type DoneFunc func()

// CreateServerFromInfra creates server with infrastructure dependencies.
// It returns created server and finalize function.
// a nil config is OK and use DefaultConfig insteadly.
func CreateServerFromInfra(repos domain.Repositories, qs *chat.Queryers, ps chat.Pubsub, conf *Config) (*Server, DoneFunc) {
	chatCmd := chat.NewCommandServiceImpl(repos, ps)
	chatQuery := chat.NewQueryServiceImpl(qs)
	chatHub := chat.NewHubImpl(chatCmd)
	go chatHub.Listen(context.Background())

	login := chat.NewLoginServiceImpl(qs.UserQueryer, ps)

	server := NewServer(chatCmd, chatQuery, chatHub, login, conf)
	doneFunc := func() {
		chatHub.Shutdown()
	}
	return server, doneFunc
}
