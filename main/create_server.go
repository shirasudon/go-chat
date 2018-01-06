package main

import (
	"context"

	gochat "github.com/shirasudon/go-chat"
	"github.com/shirasudon/go-chat/chat"
	"github.com/shirasudon/go-chat/domain"
)

// createServer creates server with resolve its dependencies.
func createServer(repos domain.Repositories, qs *chat.Queryers, ps chat.Pubsub) (server *gochat.Server, done DoneFunc) {
	chatCmd := chat.NewCommandServiceImpl(repos, ps)
	chatQuery := chat.NewQueryServiceImpl(qs)
	chatHub := chat.NewHubImpl(chatCmd)
	go chatHub.Listen(context.Background())

	done = func() {
		chatHub.Shutdown()
	}
	server = gochat.NewServer(chatCmd, chatQuery, chatHub, qs.UserQueryer, nil)
	return
}
