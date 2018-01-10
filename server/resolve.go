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
func CreateServerFromInfra(repos domain.Repositories, qs *chat.Queryers, ps chat.Pubsub) (*Server, DoneFunc) {
	chatCmd := chat.NewCommandServiceImpl(repos, ps)
	chatQuery := chat.NewQueryServiceImpl(qs)
	chatHub := chat.NewHubImpl(chatCmd)
	go chatHub.Listen(context.Background())

	server := NewServer(chatCmd, chatQuery, chatHub, qs.UserQueryer, nil)
	doneFunc := func() {
		chatHub.Shutdown()
	}
	return server, doneFunc
}