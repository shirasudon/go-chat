package main

import (
	"log"

	"github.com/mzki/chat"
)

func main() {
	s := chat.NewServer(nil)
	log.Fatal(s.ListenAndServe())
}
