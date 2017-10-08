package main

import (
	"log"

	"github.com/shirasudon/go-chat"
	"github.com/shirasudon/go-chat/domain"
	_ "github.com/shirasudon/go-chat/infra/stub"
)

func main() {
	// initilize database
	repos, err := domain.OpenRepositories("stub")
	if err != nil {
		log.Fatal(err)
	}

	log.Fatal(chat.ListenAndServe(repos, nil))
}
