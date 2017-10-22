package main

import (
	"log"

	"github.com/shirasudon/go-chat"
	"github.com/shirasudon/go-chat/infra/inmemory"
)

func main() {
	// initilize database
	repos := inmemory.OpenRepositories()
	defer repos.Close()
	log.Fatal(chat.ListenAndServe(repos, nil))
}
