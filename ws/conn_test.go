package ws

import (
	"context"
	"strings"
	"testing"

	"github.com/shirasudon/go-chat/entity"
	"github.com/shirasudon/go-chat/model/action"
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

		cm := action.ChatMessage{Content: GreetingMsg}
		cm.ActionName = action.ActionChatMessage
		conn := NewConn(ws, entity.User{})
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
	var cm action.ChatMessage
	if err := websocket.JSON.Receive(conn, &cm); err != nil {
		t.Fatalf("client receive error: %v", err)
	}

	if cm.Content != GreetingMsg {
		t.Errorf("different received message, got: %v, expect: %v", cm.Content, GreetingMsg)
	}

	// Send msg received from above.
	if err := websocket.JSON.Send(conn, cm); err != nil {
		t.Fatalf("client send error: %v", err)
	}

	// Send invalid message and Receive error message
	if err := websocket.JSON.Send(conn, "aa"); err != nil {
		t.Fatalf("client send error: %v", err)
	}

	var (
		anyMsg action.AnyMessage
		errMsg action.ErrorMessage
	)
	if err := websocket.JSON.Receive(conn, &anyMsg); err != nil {
		t.Fatalf("client receive error: %v", err)
	}

	errMsg, err = action.ParseErrorMessage(anyMsg, action.ActionError)
	if err != nil {
		t.Fatalf("parse ErrorMessage fail: %v", err)
	}
	if len(errMsg.ErrorMsg) == 0 {
		t.Errorf("got error message but message is empty")
	}
	t.Logf("LOG: send invalid message, then return: %v", errMsg.ErrorMsg)

	// Send no actiom Message
	cm = action.ChatMessage{}
	cm.ActionName = action.ActionEmpty
	if err := websocket.JSON.Send(conn, cm); err != nil {
		t.Fatalf("client send error: %v", err)
	}
	if err := websocket.JSON.Receive(conn, &anyMsg); err != nil {
		t.Fatalf("client receive error: %v", err)
	}

	errMsg, err = action.ParseErrorMessage(anyMsg, action.ActionError)
	if err != nil {
		t.Fatalf("parse ErrorMessage fail: %v", err)
	}

	if len(errMsg.ErrorMsg) == 0 {
		t.Errorf("got error message but message is empty")
	}
	t.Logf("LOG: send invalid message, then return: %v", errMsg.ErrorMsg)
}

func TestConnClose(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	endCh := make(chan bool, 1)
	server := wstest.NewServer(websocket.Handler(func(ws *websocket.Conn) {
		defer ws.Close()

		conn := NewConn(ws, entity.User{})
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
