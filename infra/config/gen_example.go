// +build ignore

package main

// this file generates example toml file into ./example/

import (
	"log"
	"os"

	"github.com/BurntSushi/toml"

	"github.com/shirasudon/go-chat/server"
)

const WriteFile = "./example/config.toml"

func main() {
	conf := server.DefaultConfig

	fp, err := os.Create(WriteFile)
	if err != nil {
		log.Fatal(err)
	}
	defer fp.Close()

	if err := toml.NewEncoder(fp).Encode(conf); err != nil {
		log.Fatal(err)
	}
	log.Println("write default config to", WriteFile)
}
