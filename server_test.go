package chat

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"golang.org/x/net/websocket"
)

func createConn(requestPath, origin string) (*websocket.Conn, error) {
	wsURL := strings.Replace(requestPath, "http://", "ws://", 1)
	return websocket.Dial(wsURL, "", origin)
}

func TestServer(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(NewServer(nil).routingRoom))
	defer ts.Close()

	requestPath := ts.URL + "/chat/ws/room1"
	origin := ts.URL[0:strings.LastIndex(ts.URL, ":")]

	conn, err := createConn(requestPath, origin)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
}
