// +build appengine

package app

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/shirasudon/go-chat/chat"
	"github.com/shirasudon/go-chat/domain"
	"github.com/shirasudon/go-chat/infra/config"
	"github.com/shirasudon/go-chat/infra/inmemory"
	"github.com/shirasudon/go-chat/infra/pubsub"
	goserver "github.com/shirasudon/go-chat/server"

	"google.golang.org/appengine"
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

func loadConfig() *goserver.Config {
	// get config path from environment value.
	var configPath = DefaultConfigFile
	if confPath := os.Getenv(KeyConfigFileENV); len(confPath) > 0 {
		configPath = confPath
	}

	// set config value to be used.
	var defaultConf = goserver.DefaultConfig
	if config.FileExists(configPath) {
		log.Printf("[Config] Loading file: %s\n", configPath)

		loaded, err := config.LoadFile(configPath)
		if err != nil {
			log.Printf("[Config] Load Error: %v\n", err)
			log.Println("[Config] use default insteadly")
			return &defaultConf
		}
		defaultConf = *loaded
		log.Println("[Config] Loading file: OK")

	} else {
		log.Println("[Config] Use default")
	}
	return &defaultConf
}

var (
	gochatServer *goserver.Server
	doneFunc     func()
)

func init() {
	var serverDoneFunc func()
	repos, qs, ps, infraDoneFunc := createInfra()
	gochatServer, serverDoneFunc = goserver.CreateServerFromInfra(repos, qs, ps, loadConfig())
	doneFunc = func() {
		serverDoneFunc()
		infraDoneFunc()
	}
	http.Handle("/", gochatServer.Handler())
}

func main() {
	defer func() {
		log.Println("calling main defer")
		doneFunc()
	}()
	appengine.Main()
}
