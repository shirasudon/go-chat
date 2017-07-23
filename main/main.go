package main

import (
	"github.com/mzki/chat"
)

func main() {
	_ = chat.NewServer(nil).ListenAndServe()
}
