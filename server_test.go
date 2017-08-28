package chat

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"golang.org/x/net/websocket"

	"github.com/labstack/echo"
	"github.com/shirasudon/go-chat/entity"
	_ "github.com/shirasudon/go-chat/entity/stub"
	"github.com/shirasudon/go-chat/model"
	"github.com/shirasudon/go-chat/wstest"
)

var (
	repository entity.Repositories
	server     *Server
)

const (
	LoginUserID = 2
)

func init() {
	repository, _ = entity.OpenRepositories("stub")
	server = NewServer(repository, nil)
}

// TODO
func TestServerServeChatWebsocket(t *testing.T) {
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
	defer server.Shutdown(context.Background())

	requestPath := ts.URL + "/chat/ws"
	origin := ts.URL[0:strings.LastIndex(ts.URL, ":")]

	// create websocket connection for testiong server ts.
	conn, err := wstest.NewClientConn(requestPath, origin)
	if err != nil {
		t.Fatalf("can not create websocket connetion, error: %v", err)
	}
	defer conn.Close()

	// write message to server
	writeCM := model.ChatMessage{Content: "hello!"}
	writeCM.RoomID = 3
	writeCM.ActionName = model.ActionChatMessage
	if err := websocket.JSON.Send(conn, writeCM); err != nil {
		t.Fatal(err)
	}

	// read message from server
	var readAny model.AnyMessage
	if err := websocket.JSON.Receive(conn, &readAny); err != nil {
		t.Fatal(err)
	}
	if readAny.Action() != model.ActionChatMessage {
		t.Fatalf("%#v", readAny)
	}
	readCM := model.ParseChatMessage(readAny, readAny.Action())

	// check same message
	if readCM.Content != writeCM.Content {
		t.Errorf("different chat message, got: %#v, expect: %#v", readCM, writeCM)
	}
}
