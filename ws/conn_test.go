package ws

import (
	"context"
	"strings"
	"testing"

	"github.com/shirasudon/go-chat/chat/action"
	"github.com/shirasudon/go-chat/domain"
	"github.com/shirasudon/go-chat/ws/wstest"

	"golang.org/x/net/websocket"
)

const GreetingMsg = "hello!"

func TestNewConn(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	endCh := make(chan bool, 1)
	server := wstest.NewServer(websocket.Handler(func(ws *websocket.Conn) {
		defer ws.Close()

		cm := domain.MessageCreated{Content: GreetingMsg}
		conn := NewConn(ws, domain.User{})
		conn.Send(cm)
		conn.Listen(ctx)
		endCh <- true
	}))
	defer func() {
		server.Close()
		<-endCh
	}()

	requestPath := server.URL + "/ws"
	origin := server.URL[0:strings.LastIndex(server.URL, ":")]
	conn, err := wstest.NewClientConn(requestPath, origin)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	// Receive hello message
	var created domain.MessageCreated
	if err := websocket.JSON.Receive(conn, &created); err != nil {
		t.Fatalf("client receive error: %v", err)
	}

	if created.Content != GreetingMsg {
		t.Errorf("different received message, got: %v, expect: %v", created.Content, GreetingMsg)
	}

	// Send msg received from above.
	var cm = action.ChatMessage{Content: created.Content}
	cm.ActionName = action.ActionChatMessage
	if err := websocket.JSON.Send(conn, cm); err != nil {
		t.Fatalf("client send error: %v", err)
	}

	// Send invalid message and Receive error message
	if err := websocket.JSON.Send(conn, "aa"); err != nil {
		t.Fatalf("client send error: %v", err)
	}

	var (
		anyMsg map[string]interface{}
	)
	if err := websocket.JSON.Receive(conn, &anyMsg); err != nil {
		t.Fatalf("client receive error: %v", err)
	}

	errMsg, ok := anyMsg["message"].(string)
	if !ok {
		t.Fatalf("got invalid error message: %#v", anyMsg)
	}
	if len(errMsg) == 0 {
		t.Errorf("got error message but message is empty")
	}
	t.Logf("LOG: send invalid message, then return: %v", errMsg)

	// Send no action Message
	cm = action.ChatMessage{}
	cm.ActionName = action.ActionEmpty
	if err := websocket.JSON.Send(conn, cm); err != nil {
		t.Fatalf("client send error: %v", err)
	}
	if err := websocket.JSON.Receive(conn, &anyMsg); err != nil {
		t.Fatalf("client receive error: %v", err)
	}

	errMsg, ok = anyMsg["message"].(string)
	if !ok {
		t.Fatalf("got invalid error message: %#v", anyMsg)
	}
	if len(errMsg) == 0 {
		t.Errorf("got error message but message is empty")
	}
	t.Logf("LOG: send invalid message, then return: %v", errMsg)
}

func TestConnClose(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	endCh := make(chan bool, 1)
	server := wstest.NewServer(websocket.Handler(func(ws *websocket.Conn) {
		defer ws.Close()

		conn := NewConn(ws, domain.User{})
		conn.Close() // to quit Listen() immediately
		conn.Listen(ctx)
		endCh <- true
	}))
	defer func() {
		server.Close()
		<-endCh
	}()

	requestPath := server.URL + "/ws"
	origin := server.URL[0:strings.LastIndex(server.URL, ":")]
	conn, err := wstest.NewClientConn(requestPath, origin)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
}
