package chat

import (
	"context"
	"errors"
	"log"
	"strings"

	"golang.org/x/net/websocket"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/shirasudon/go-chat/chat"
	"github.com/shirasudon/go-chat/chat/action"
	"github.com/shirasudon/go-chat/domain"
	"github.com/shirasudon/go-chat/infra/pubsub"
	"github.com/shirasudon/go-chat/ws"
)

// it represents server which can accepts chat room and its clients.
type Server struct {
	echo            *echo.Echo
	websocketServer *websocket.Server
	loginHandler    *LoginHandler
	restHandler     *RESTHandler

	chatHub   *chat.Hub
	chatCmd   *chat.CommandService
	chatQuery *chat.QueryService
	pubsub    *pubsub.PubSub

	repos domain.Repositories

	conf Config
}

// it returns new constructed server with config.
// nil config is ok and use DefaultConfig insteadly.
func NewServer(repos domain.Repositories, conf *Config) *Server {
	if conf == nil {
		conf = &DefaultConfig
	}

	e := echo.New()
	e.HideBanner = true

	pubsub := pubsub.New(10)
	chatCmd := chat.NewCommandService(repos, pubsub)
	chatQuery := chat.NewQueryService(repos)

	s := &Server{
		echo:         e,
		loginHandler: NewLoginHandler(repos.Users()),
		restHandler:  NewRESTHandler(chatCmd, chatQuery),
		chatHub:      chat.NewHub(chatCmd, chatQuery),
		chatCmd:      chatCmd,
		chatQuery:    chatQuery,
		pubsub:       pubsub,

		repos: repos,
		conf:  *conf,
	}
	s.websocketServer = &websocket.Server{} // TODO it is needed?
	return s
}

func (s *Server) serveChatWebsocket(c echo.Context) error {
	// LoggedInUserID is valid at middleware layer, loginHandler.Filter.
	userID, ok := LoggedInUserID(c)
	if !ok {
		return errors.New("needs logged in, but access without logged in state")
	}
	user, err := s.repos.Users().Find(c.Request().Context(), userID)
	if err != nil {
		return err
	}

	websocket.Handler(func(wsConn *websocket.Conn) {
		log.Println("Server.acceptWSConn: ")
		defer wsConn.Close()

		conn := ws.NewConn(wsConn, user)
		conn.OnActionMessage(func(conn *ws.Conn, m action.ActionMessage) {
			s.chatHub.Send(conn, m)
		})
		conn.OnError(func(conn *ws.Conn, err error) {
			conn.Send(action.NewErrorMessage(err))
		})
		conn.OnClosed(func(conn *ws.Conn) {
			s.chatHub.Disconnect(conn)
		})
		s.chatHub.Connect(conn)

		// blocking to avoid connection closed
		conn.Listen(c.Request().Context())
	}).ServeHTTP(c.Response(), c.Request())
	return nil
}

// it starts server process.
// it blocks until process occurs any error and
// return the error.
func (s *Server) ListenAndServe() error {
	if err := s.conf.validate(); err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// start chat process
	go s.chatHub.Listen(ctx)

	// initilize router
	e := s.echo

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// set login handler
	e.Use(s.loginHandler.Middleware())
	e.POST("/login", s.loginHandler.Login).
		Name = "doLogin"
	e.GET("/login", s.loginHandler.GetLoginState).
		Name = "getLoginInfo"
	e.POST("/logout", s.loginHandler.Logout).
		Name = "doLogout"

	chatPath := strings.TrimSuffix(s.conf.ChatPath, "/")
	chatGroup := e.Group(chatPath, s.loginHandler.Filter())

	// set restHandler
	chatGroup.POST("/users/:id/rooms", s.restHandler.CreateRoom).
		Name = "chat.createUserRoom"
	chatGroup.DELETE("/users/:id/rooms", s.restHandler.DeleteRoom).
		Name = "chat.deleteUserRoom"

	// set websocket handler
	chatGroup.GET("/ws", s.serveChatWebsocket).
		Name = "chat.connentWebsocket"

	// serve static content
	e.Static("/", "").
		Name = "staticContents"

	// show registered URLs
	for _, url := range e.Routes() {
		// built-in routes are ignored
		if strings.Contains(url.Name, "github.com/labstack/echo") {
			continue
		}
		log.Printf("%8s : %-35s (%v)\n", url.Method, url.Path, url.Name)
	}

	// start server
	serverURL := s.conf.HTTP
	log.Println("server listen at " + serverURL)
	err := e.Start(serverURL)
	e.Logger.Error(err)
	return err
}

func (s *Server) Shutdown(ctx context.Context) error {
	s.pubsub.Shutdown()
	return s.echo.Shutdown(ctx)
}

// It starts server process using default server with
// user config.
// A nil config is OK and use DefaultConfig insteadly.
// It blocks until the process occurs any error and
// return the error.
func ListenAndServe(repos domain.Repositories, conf *Config) error {
	if conf == nil {
		conf = &DefaultConfig
	}
	s := NewServer(repos, conf)
	defer s.Shutdown(context.Background())
	return s.ListenAndServe()
}
