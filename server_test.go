package chat

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/mzki/go-chat/entity"
	_ "github.com/mzki/go-chat/entity/stub"
	"github.com/mzki/go-chat/model"
	"golang.org/x/net/websocket"
)

// create client-side conncetion for the websocket
func createConn(requestPath, origin string) (*websocket.Conn, error) {
	wsURL := strings.Replace(requestPath, "http://", "ws://", 1)
	return websocket.Dial(wsURL, "", origin)
}

var repository entity.Repositories

func init() {
	repository, _ = entity.OpenRepositories("stub")
}

func TestServer(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(NewServer(repository, nil).routingRoom))
	defer ts.Close()

	requestPath := ts.URL + "/chat/ws/room1"
	origin := ts.URL[0:strings.LastIndex(ts.URL, ":")]

	conn, err := createConn(requestPath, origin)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	errCh := make(chan error)
	go func() {
		defer close(errCh)
		// write to server
		writeCM := model.ChatMessage{}
		writeCM.ActionName = model.ActionChatMessage
		if err := websocket.JSON.Send(conn, &writeCM); err != nil {
			errCh <- err
			return
		}
		// read from server
		var readCM model.ChatMessage
		if err := websocket.JSON.Receive(conn, &readCM); err != nil {
			errCh <- err
			return
		}

		// check same data
		if readCM != writeCM {
			errCh <- errors.New("different chat message")
		}
	}()

	// set timeout for this test
	timer := time.NewTimer(200 * time.Millisecond)
	defer timer.Stop()

	for {
		select {
		case err, ok := <-errCh:
			if !ok {
				return
			}
			t.Error(err)
		case <-timer.C:
			t.Error("testing timeout, please check error log")
			return
		}
	}
}
