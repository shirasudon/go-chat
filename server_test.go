package chat

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo"
	"github.com/shirasudon/go-chat/entity"
	_ "github.com/shirasudon/go-chat/entity/stub"

	"golang.org/x/net/websocket"
)

// create client-side conncetion for the websocket
func createConn(requestPath, origin string) (*websocket.Conn, error) {
	wsURL := strings.Replace(requestPath, "http://", "ws://", 1)
	return websocket.Dial(wsURL, "", origin)
}

var (
	repository entity.Repositories
	server     *Server
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
			c.Set(KeyLoggedInUserID, uint64(0)) // To use check for login state
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

	requestPath := ts.URL + "/ws/chat/ws/room1"
	origin := ts.URL[0:strings.LastIndex(ts.URL, ":")]

	// create websocket connection for testiong server ts.
	conn, err := createConn(requestPath, origin)
	if err != nil {
		t.Fatalf("can not create websocket connetion, error: %v", err)
	}
	defer conn.Close()

	// // TODO describe behaviors of the websocket client.

	// firstly enter any rooms
	// TODO

	// write message to server
	// writeCM := model.ChatMessage{}
	// writeCM.ActionName = model.ActionChatMessage
	// if err := websocket.JSON.Send(conn, writeCM); err != nil {
	// 	t.Fatal(err)
	// }

	// read message from server
	// var readCM model.ChatMessage
	// if err := websocket.JSON.Receive(conn, &readCM); err != nil {
	// 	t.Fatal(err)
	// }

	// check same message
	// if readCM != writeCM {
	// 	t.Errorf("different chat message, got: %v, expect: %v", readCM, writeCM)
	// }
}
