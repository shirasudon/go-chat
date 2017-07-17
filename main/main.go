package main

import (
	"github.com/mzki/chat"
	"log"
)

func main() {
	s := chat.NewServer(nil)
	log.Fatal(s.ListenAndServe())
}
