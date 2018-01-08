// package resolve provide helper functions for resolving dependencies to
// construct gochat Server.

package resolve

import (
	"context"
	"fmt"

	gochat "github.com/shirasudon/go-chat"
	"github.com/shirasudon/go-chat/chat"
	"github.com/shirasudon/go-chat/domain"
)

func main() {
	fmt.Println("vim-go")
}

// DoneFunc is function to be called after all of operations are done.
type DoneFunc func()

// CreateServer creates server with resolve its dependencies.
func CreateServer(repos domain.Repositories, qs *chat.Queryers, ps chat.Pubsub) (server *gochat.Server, done DoneFunc) {
	chatCmd := chat.NewCommandServiceImpl(repos, ps)
	chatQuery := chat.NewQueryServiceImpl(qs)
	chatHub := chat.NewHubImpl(chatCmd)
	go chatHub.Listen(context.Background())

	done = func() {
		chatHub.Shutdown()
	}
	server = gochat.NewServer(chatCmd, chatQuery, chatHub, qs.UserQueryer, nil)
	return
}
