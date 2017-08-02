// +build ignore

package main

import (
	"github.com/mzki/go-chat/entity/sqlite3"
	"log"
	"os"
)

const DBFile = "_test.sqlite3"

func main() {
	if err := os.Remove(DBFile); err != nil {
		log.Fatal(err)
	}

	repos, err := sqlite3.RepositoryProducer(DBFile)
	if err != nil {
		log.Fatal(err)
	}
	defer repos.Close()

	DB := repos.(*sqlite3.Repositories).DB
	if _, err := DB.Exec(`CREATE TABLE users (
		"id" INTEGER PRIMARY_KEY AUTOINCREMENT,
    "name" VARCHAR(100),
		"password" VARCHAR(100),
	)`); err != nil {
		log.Fatal(err)
	}
}
