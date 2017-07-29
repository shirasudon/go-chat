package main

import (
	"log"

	"github.com/mzki/chat"
	"github.com/mzki/chat/entity"
	_ "github.com/mzki/chat/entity/stub"
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
