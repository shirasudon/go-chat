package main

import (
	"context"
	"log"

	"github.com/shirasudon/go-chat/chat"
	"github.com/shirasudon/go-chat/domain"
	"github.com/shirasudon/go-chat/infra/inmemory"
	"github.com/shirasudon/go-chat/infra/pubsub"
	"github.com/shirasudon/go-chat/server"
)

type DoneFunc func()

func createInfra() (domain.Repositories, *chat.Queryers, chat.Pubsub, DoneFunc) {
	ps := pubsub.New()
	doneFuncs := make([]func(), 0, 4)
	doneFuncs = append(doneFuncs, ps.Shutdown)

	repos := inmemory.OpenRepositories(ps)
	doneFuncs = append(doneFuncs, func() { _ = repos.Close() })

	ctx, cancel := context.WithCancel(context.Background())
	doneFuncs = append(doneFuncs, cancel)
	go repos.UpdatingService(ctx)

	qs := &chat.Queryers{
		UserQueryer:    repos.UserRepository,
		RoomQueryer:    repos.RoomRepository,
		MessageQueryer: repos.MessageRepository,
		EventQueryer:   repos.EventRepository,
	}

	done := func() {
		// reverse order to simulate defer statement.
		for i := len(doneFuncs); i >= 0; i-- {
			doneFuncs[i]()
		}
	}

	return repos, qs, ps, done
}

func main() {
	repos, qs, ps, infraDoneFunc := createInfra()
	s, done := server.CreateServerFromInfra(repos, qs, ps)
	defer func() {
		done()
		infraDoneFunc()
	}()

	if err := s.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
