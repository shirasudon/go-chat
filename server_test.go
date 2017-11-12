package chat

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"golang.org/x/net/websocket"

	"github.com/labstack/echo"
	"github.com/shirasudon/go-chat/chat"
	"github.com/shirasudon/go-chat/chat/action"
	"github.com/shirasudon/go-chat/infra/inmemory"
	"github.com/shirasudon/go-chat/ws/wstest"
)

var (
	repository = inmemory.OpenRepositories()

	queryers *chat.Queryers = &chat.Queryers{
		UserQueryer:    repository.UserRepository,
		RoomQueryer:    repository.RoomRepository,
		MessageQueryer: repository.MessageRepository,
	}
)

const (
	LoginUserID = 2
)

func TestServerServeChatWebsocket(t *testing.T) {
	server := NewServer(repository, queryers, nil)
	defer server.Shutdown(context.Background())

	e := echo.New()
	serverErrCh := make(chan error, 1)
	ts := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			c := e.NewContext(req, w)
			c.Set(KeyLoggedInUserID, uint64(LoginUserID)) // To use check for login state
			if err := server.serveChatWebsocket(c); err != nil {
				serverErrCh <- err
			}
		}),
	)
	defer func() {
		ts.Close()
		// catch server error at end
		select {
		case err := <-serverErrCh:
			if err != nil {
				t.Errorf("request handler returns erorr: %v", err)
			}
		default:
		}
	}()

	// run server process
	go func() {
		server.ListenAndServe()
	}()

	requestPath := ts.URL + "/chat/ws"
	origin := ts.URL[0:strings.LastIndex(ts.URL, ":")]

	// create websocket connection for testiong server ts.
	conn, err := wstest.NewClientConn(requestPath, origin)
	if err != nil {
		t.Fatalf("can not create websocket connetion, error: %v", err)
	}
	defer conn.Close()

	// write message to server
	cm := action.ChatMessage{Content: "hello!"}
	cm.RoomID = 3
	toSend := map[string]interface{}{
		action.KeyAction: action.ActionChatMessage,
		"data":           cm,
	}
	if err := websocket.JSON.Send(conn, toSend); err != nil {
		t.Fatal(err)
	}

	// read message from server
	var readAny map[string]interface{}
	if err := websocket.JSON.Receive(conn, &readAny); err != nil {
		t.Fatal(err)
	}
	if got, expect := readAny["event"], chat.EventNameMessageCreated; got != expect {
		t.Errorf("diffrent event names, expect: %v, got: %v", expect, got)
	}
	created, ok := readAny["data"].(map[string]interface{})
	if !ok {
		t.Fatalf("invalid data is recieved: %#v", readAny)
	}

	// check same message
	if created["content"].(string) != cm.Content {
		t.Errorf("different chat message fields, recieved: %#v, send: %#v", created, toSend)
	}
}
