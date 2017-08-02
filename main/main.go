package main

import (
	"log"

	"github.com/shirasudon/go-chat"
	"github.com/shirasudon/go-chat/entity"
	_ "github.com/shirasudon/go-chat/entity/stub"
)

func main() {
	// initilize database
	repos, err := entity.OpenRepositories("stub")
	if err != nil {
		log.Fatal(err)
	}
	defer repos.Close()

	log.Fatal(chat.ListenAndServe(repos, nil))
}
