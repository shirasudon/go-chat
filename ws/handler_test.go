package ws

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/shirasudon/go-chat/domain/event"
	"github.com/shirasudon/go-chat/ws/wstest"
	"golang.org/x/net/websocket"
)

func TestNewServerFunc(t *testing.T) {
	t.Parallel()

	// succeed
	_ = NewServerFunc(func(c *Conn) {
		// do nothing
	})

	// panic
	defer func() {
		if rec := recover(); rec == nil {
			t.Error("expect panic is occured but no panic")
		}
	}()
	_ = NewServerFunc(nil)
}

func TestServeHTTPWithUserID(t *testing.T) {
	const (
		UserID   = uint64(2)
		WaitTime = 20 * time.Millisecond
	)

	var (
		done    = make(chan bool, 1)
		timeout = time.After(WaitTime)

		HandlerPassed bool = false
	)

	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		s := NewServerFunc(func(c *Conn) {
			if c.UserID() != UserID {
				t.Errorf("different user ID in the connection, expect: %v, got: %v", UserID, c.UserID())
			}
			HandlerPassed = true
		})
		s.ServeHTTPWithUserID(w, req, UserID)
		if !HandlerPassed {
			t.Error("Handler function is not called")
		}
		done <- true
	}))

	defer func() {
		testServer.Close()
		select {
		case <-done:
		case <-timeout:
			t.Error("testing is timeouted")
		}
	}()

	requestPath := testServer.URL + "/ws"
	origin := testServer.URL[0:strings.LastIndex(testServer.URL, ":")]
	conn, err := wstest.NewClientConn(requestPath, origin)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
}

func TestWsHandlerFail(t *testing.T) {
	const (
		UserID   = uint64(2)
		WaitTime = 20 * time.Millisecond
	)

	var (
		done    = make(chan bool, 1)
		timeout = time.After(WaitTime)

		HandlerPassed bool = false
	)

	testServer := wstest.NewServer(websocket.Handler(func(wsConn *websocket.Conn) {
		s := NewServerFunc(func(c *Conn) { HandlerPassed = true })
		s.wsHandler(wsConn)
		if HandlerPassed {
			t.Error("Handler function is expected to never called but called")
		}
		done <- true
	}))

	defer func() {
		testServer.Close()
		select {
		case <-done:
		case <-timeout:
			t.Error("testing is timeouted")
		}
	}()

	requestPath := testServer.URL + "/ws"
	origin := testServer.URL[0:strings.LastIndex(testServer.URL, ":")]
	conn, err := wstest.NewClientConn(requestPath, origin)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	var errRaised = event.ErrorRaised{}
	conn.SetDeadline(time.Now().Add(WaitTime))
	if err := websocket.JSON.Receive(conn, &errRaised); err != nil {
		t.Fatalf("websocket receiving fail: %v", err)
	}
	if len(errRaised.Message) == 0 {
		t.Errorf("error message should not be empty")
	}
}

func TestGetSetConnectUserID(t *testing.T) {
	t.Parallel()

	const (
		UserID = uint64(3)
	)

	ctx := context.Background()

	// can not get userID from empty context.
	_, err := getConnectUserID(ctx)
	if err == nil {
		t.Fatal("got userID from empty context")
	}

	newCtx := setConnectUserID(ctx, UserID)

	// can get userID from new context.
	got, err := getConnectUserID(newCtx)
	if err != nil {
		t.Fatalf("can not get userID from the context set by setConnectUserID, err: %v", err)
	}
	if got != UserID {
		t.Errorf("different user ID in the context, expect: %v, got: %v", UserID, got)
	}
}
