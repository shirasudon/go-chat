// +build appengine

package main

import (
	"context"
	"net/http"

	gochat "github.com/shirasudon/go-chat"
	"github.com/shirasudon/go-chat/chat"
	"github.com/shirasudon/go-chat/domain"
	"github.com/shirasudon/go-chat/infra/inmemory"
	"github.com/shirasudon/go-chat/infra/pubsub"

	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
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

var (
	gochatServer *gochat.Server
	doneFunc     func()
)

func init() {
	var serverDoneFunc func()
	repos, qs, ps, infraDoneFunc := createInfra()
	gochatServer, serverDoneFunc = createServer(repos, qs, ps)
	doneFunc = func() {
		serverDoneFunc()
		infraDoneFunc()
	}

	http.Handle("/", gochatServer.Handler())
}

func main() {
	defer func() {
		doneFunc()
		log.Debugf(context.Background(), "calling main defer")
	}()
	appengine.Main()
}
