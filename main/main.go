package main

import (
	"context"
	"log"
	"os"

	"github.com/shirasudon/go-chat/chat"
	"github.com/shirasudon/go-chat/domain"
	"github.com/shirasudon/go-chat/infra/config"
	"github.com/shirasudon/go-chat/infra/inmemory"
	"github.com/shirasudon/go-chat/infra/pubsub"
	"github.com/shirasudon/go-chat/server"
)

type DoneFunc func()

func createInfra() (domain.Repositories, *chat.Queryers, chat.Pubsub, DoneFunc) {
	ps := pubsub.New()
	doneFuncs := make([]func(), 0, 4)
	doneFuncs = append(doneFuncs, ps.Shutdown)

	repos := inmemory.OpenRepositories(ps)
	doneFuncs = append(doneFuncs, func() { _ = repos.Close() })

	ctx, cancel := context.WithCancel(context.Background())
	doneFuncs = append(doneFuncs, cancel)
	go repos.UpdatingService(ctx)

	qs := &chat.Queryers{
		UserQueryer:    repos.UserRepository,
		RoomQueryer:    repos.RoomRepository,
		MessageQueryer: repos.MessageRepository,
		EventQueryer:   repos.EventRepository,
	}

	done := func() {
		// reverse order to simulate defer statement.
		for i := len(doneFuncs); i >= 0; i-- {
			doneFuncs[i]()
		}
	}

	return repos, qs, ps, done
}

const (
	DefaultConfigFile = "config.toml"
	KeyConfigFileENV  = "GOCHAT_CONFIG_FILE"
)

func main() {
	// get config path from environment value.
	var configPath = DefaultConfigFile
	if confPath := os.Getenv(KeyConfigFileENV); len(confPath) > 0 {
		configPath = confPath
	}

	// set config value to be used.
	var defaultConf = server.DefaultConfig
	if config.FileExists(configPath) {
		log.Printf("[Config] Loading file: %s\n", configPath)
		if err := config.LoadFile(&defaultConf, configPath); err != nil {
			log.Fatalf("[Config] Load Error: %v", err)
		}
		log.Println("[Config] Loading file: OK")
	} else {
		log.Println("[Config] Use default")
	}

	repos, qs, ps, infraDoneFunc := createInfra()
	s, done := server.CreateServerFromInfra(repos, qs, ps, &defaultConf)
	defer func() {
		done()
		infraDoneFunc()
	}()

	if err := s.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
	log.Println("[Server] quiting...")
}
