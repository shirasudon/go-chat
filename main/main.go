package main

import (
	"log"

	"github.com/shirasudon/go-chat"
	"github.com/shirasudon/go-chat/infra/stub"
)

func main() {
	// initilize database
	repos := stub.OpenRepositories()
	defer repos.Close()
	log.Fatal(chat.ListenAndServe(repos, nil))
}
