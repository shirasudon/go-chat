package server

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"golang.org/x/net/websocket"

	"github.com/labstack/echo"

	"github.com/shirasudon/go-chat/chat"
	"github.com/shirasudon/go-chat/chat/action"
	"github.com/shirasudon/go-chat/domain"
	"github.com/shirasudon/go-chat/infra/inmemory"
	"github.com/shirasudon/go-chat/infra/pubsub"
	"github.com/shirasudon/go-chat/ws/wstest"
)

var (
	globalPubsub = pubsub.New()
	repository   = inmemory.OpenRepositories(globalPubsub)

	queryers *chat.Queryers = &chat.Queryers{
		UserQueryer:    repository.UserRepository,
		RoomQueryer:    repository.RoomRepository,
		MessageQueryer: repository.MessageRepository,
		EventQueryer:   repository.EventRepository,
	}

	chatCmd   = chat.NewCommandServiceImpl(repository, globalPubsub)
	chatQuery = chat.NewQueryServiceImpl(queryers)
	chatHub   = chat.NewHubImpl(chatCmd)

	loginService = chat.NewLoginServiceImpl(queryers.UserQueryer, globalPubsub)

	theEcho = echo.New()
)

func TestMain(m *testing.M) {
	defer globalPubsub.Shutdown()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go repository.UpdatingService(ctx)

	go chatHub.Listen(ctx)
	defer chatHub.Shutdown()

	time.Sleep(1 * time.Millisecond) // wait for the starting of UpdatingServices.

	os.Exit(m.Run())
}

const (
	LoginUserID = 2
)

func TestServerListenAndServeFailWithConfig(t *testing.T) {
	var Config = Config{
		HTTP: "unknown",
	}
	if err := Config.Validate(); err == nil {
		t.Fatal("Config should be invalid here")
	}

	server := NewServer(chatCmd, chatQuery, chatHub, loginService, &Config)
	defer server.Shutdown(context.Background())

	errCh := make(chan error, 1)
	timeout := time.After(1 * time.Millisecond)
	go func() {
		errCh <- server.ListenAndServe()
	}()
	select {
	case err := <-errCh:
		if err == nil {
			t.Errorf("the ListenAndServe should return error by invalid config")
		}
		// occuring error is OK
	case <-timeout:
		t.Errorf("timeout")
	}
}

func TestServerServeChatWebsocket(t *testing.T) {
	server := NewServer(chatCmd, chatQuery, chatHub, loginService, nil)
	defer server.Shutdown(context.Background())

	// run server process
	go func() {
		server.ListenAndServe()
	}()

	// waiting for the server process stands up.
	time.Sleep(10 * time.Millisecond)

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

	requestPath := ts.URL + "/chat/ws"
	origin := ts.URL[0:strings.LastIndex(ts.URL, ":")]

	// create websocket connection for testiong server ts.
	conn, err := wstest.NewClientConn(requestPath, origin)
	if err != nil {
		t.Fatalf("can not create websocket connetion, error: %v", err)
	}
	defer conn.Close()

	// read connected event from server.
	// to avoid infinite loops, we set read dead line to 100ms.
	conn.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
	{
		var readAny map[string]interface{}
		if err := websocket.JSON.Receive(conn, &readAny); err != nil {
			t.Fatal(err)
		}

		if got, expect := readAny["event"], chat.EventNameActiveClientActivated; got != expect {
			t.Errorf("diffrent event names, expect: %v, got: %v", expect, got)
		}

		activated, ok := readAny["data"].(map[string]interface{})
		if !ok {
			t.Fatalf("invalid data is recieved: %#v", readAny)
		}
		if got := activated["user_id"].(float64); got != LoginUserID {
			t.Errorf("diffrent user id, expect: %v, got: %v", LoginUserID, got)
		}
	}

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

	// read message from server.
	// to avoid infinite loops, we set read dead line to 100ms.
	conn.SetReadDeadline(time.Now().Add(100 * time.Millisecond))

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

func TestServerConnIsClosedAfterLogout(t *testing.T) {
	// setup user to be used here
	var (
		err      error = nil
		testUser       = domain.User{Name: "test_user", Password: "password"}
	)
	testUser.ID, err = repository.UserRepository.Store(context.Background(), testUser)
	if err != nil {
		t.Fatal(err)
	}

	server := NewServer(chatCmd, chatQuery, chatHub, loginService, nil)
	defer server.Shutdown(context.Background())

	// run server process
	go func() {
		server.ListenAndServe()
	}()

	// waiting for the server process stands up.
	time.Sleep(10 * time.Millisecond)

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

	requestPath := ts.URL + "/chat/ws"
	origin := ts.URL[0:strings.LastIndex(ts.URL, ":")]

	// create websocket connection for testiong server ts.
	conn, err := wstest.NewClientConn(requestPath, origin)
	if err != nil {
		t.Fatalf("can not create websocket connetion, error: %v", err)
	}
	defer conn.Close()

	// login by using login_test.doLogin.
	loginC, err := doLogin(server.loginHandler, testUser.Name, testUser.Password, false)
	if err != nil {
		t.Fatal(err)
	}
	// logout by using login_test.doLogout
	_, err = doLogout(server.loginHandler, loginC.Response().Header()["Set-Cookie"])
	if err != nil {
		t.Fatal(err)
	}

	// check whether conn is closed after logout.
	conn.SetDeadline(time.Now().Add(50 * time.Millisecond))
	_, err = conn.Read(make([]byte, 0, 32))
	if err, ok := err.(net.Error); ok && err.Timeout() {
		t.Fatal("conn is not closed after logout and timeout-ed")
	}
	if err != nil {
		t.Logf("got error :%#v", err)
	}
	// PASS
}

func TestServerHandler(t *testing.T) {
	server := NewServer(chatCmd, chatQuery, chatHub, loginService, nil)
	defer server.Shutdown(context.Background())

	// check type
	var h http.Handler = server.Handler()
	if h == nil {
		t.Fatal("Server.Handler returns nil")
	}
}
