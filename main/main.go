package main

import (
	"log"

	gochat "github.com/shirasudon/go-chat"
	"github.com/shirasudon/go-chat/chat"
	"github.com/shirasudon/go-chat/infra/inmemory"
)

func main() {
	// initilize database
	repos := inmemory.OpenRepositories()
	defer repos.Close()

	qs := &chat.Queryers{
		UserQueryer:    repos.UserRepository,
		RoomQueryer:    repos.RoomRepository,
		MessageQueryer: repos.MessageRepository,
	}
	log.Fatal(gochat.ListenAndServe(repos, qs, nil))
}
