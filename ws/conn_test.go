package ws

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/shirasudon/go-chat/chat/action"
	"github.com/shirasudon/go-chat/domain/event"
	"github.com/shirasudon/go-chat/ws/wstest"

	"golang.org/x/net/websocket"
)

const GreetingMsg = "hello!"
const Timeout = 10 * time.Millisecond

func TestNewConn(t *testing.T) {
	const (
		UserID = uint64(1)
	)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	endCh := make(chan bool, 1)
	server := wstest.NewServer(websocket.Handler(func(ws *websocket.Conn) {
		defer ws.Close()

		cm := event.MessageCreated{Content: GreetingMsg}
		conn := NewConn(ws, UserID)
		conn.Send(cm)
		conn.OnActionMessage(func(conn *Conn, m action.ActionMessage) {
			cm, ok := m.(action.ChatMessage)
			if !ok {
				t.Fatalf("invalid message structure: %#v", m)
			}
			if cm.Content != GreetingMsg {
				t.Errorf("invalid message content, expect: %v, got: %v", GreetingMsg, cm.Content)
			}
			endCh <- true // PASS this test.
		})
		conn.OnError(func(conn *Conn, err error) {
			t.Fatalf("server side conn got error: %v", err)
		})
		conn.Listen(ctx)
	}))
	defer func() {
		server.Close()

		select {
		case <-endCh:
		case <-time.After(Timeout):
			t.Error("Timeouted")
		}
	}()

	requestPath := server.URL + "/ws"
	origin := server.URL[0:strings.LastIndex(server.URL, ":")]
	conn, err := wstest.NewClientConn(requestPath, origin)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	// Receive hello message
	var created event.MessageCreated
	if err := websocket.JSON.Receive(conn, &created); err != nil {
		t.Fatalf("client receive error: %v", err)
	}

	if created.Content != GreetingMsg {
		t.Errorf("different received message, got: %v, expect: %v", created.Content, GreetingMsg)
	}

	// Send msg received from above.
	var toSend = map[string]interface{}{
		action.KeyAction: action.ActionChatMessage,
		"data":           action.ChatMessage{Content: created.Content},
	}
	if err := websocket.JSON.Send(conn, toSend); err != nil {
		t.Fatalf("client send error: %v", err)
	}
}

func TestConnGotsInvalidMessages(t *testing.T) {
	const (
		UserID = uint64(1)
	)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	endCh := make(chan bool, 1)
	server := wstest.NewServer(websocket.Handler(func(ws *websocket.Conn) {
		defer ws.Close()

		conn := NewConn(ws, UserID)
		conn.OnActionMessage(func(conn *Conn, m action.ActionMessage) {
			t.Fatalf("In this test, server side conn will never get message, but got: %#v", m)
		})
		conn.OnError(func(conn *Conn, err error) {
			t.Logf("In this test, server side conn will get error: %v", err)
		})
		conn.Listen(ctx)
		endCh <- true
	}))
	defer func() {
		server.Close()
		select {
		case <-endCh:
		case <-time.After(Timeout):
			t.Error("Timeouted")
		}
	}()

	requestPath := server.URL + "/ws"
	origin := server.URL[0:strings.LastIndex(server.URL, ":")]
	conn, err := wstest.NewClientConn(requestPath, origin)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

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
	cm := action.ChatMessage{}
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
	const (
		UserID = uint64(1)
	)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	endCh := make(chan bool, 1)
	server := wstest.NewServer(websocket.Handler(func(ws *websocket.Conn) {
		defer ws.Close()

		conn := NewConn(ws, UserID)
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
