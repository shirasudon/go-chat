package chat

import (
	"context"
	"errors"
	"log"
	"strings"

	"golang.org/x/net/websocket"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/shirasudon/go-chat/domain"
	"github.com/shirasudon/go-chat/model"
	"github.com/shirasudon/go-chat/model/action"
	"github.com/shirasudon/go-chat/ws"
)

// it represents server which can accepts chat room and its clients.
type Server struct {
	echo            *echo.Echo
	websocketServer *websocket.Server
	loginHandler    *LoginHandler
	chatHub         *model.ChatHub

	ctx context.Context

	repos domain.Repositories

	conf Config
}

// it returns new constructed server with config.
// nil config is ok and use DefaultConfig insteadly.
func NewServer(repos domain.Repositories, conf *Config) *Server {
	if conf == nil {
		conf = &DefaultConfig
	}

	s := &Server{
		loginHandler: NewLoginHandler(repos.Users()),
		chatHub:      model.NewChatHub(repos),
		repos:        repos,
		conf:         *conf,
	}
	s.websocketServer = &websocket.Server{} // TODO it is needed?
	return s
}

func (s *Server) serveChatWebsocket(c echo.Context) error {
	// KeyLoggedInUserID is set at middleware layer, loginHandler.Filter.
	userID, ok := c.Get(KeyLoggedInUserID).(uint64)
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
	e := echo.New()
	e.HideBanner = true
	s.echo = e

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// set login handler
	e.Use(s.loginHandler.Middleware())
	e.POST("/login", s.loginHandler.Login)
	e.GET("/login", s.loginHandler.GetLoginState)
	e.POST("/logout", s.loginHandler.Logout)

	// set websocket handler
	{
		chatPath := strings.TrimSuffix(s.conf.ChatPath, "/")
		chatGroup := e.Group(chatPath, s.loginHandler.Filter())
		wsRoute := chatGroup.GET("/ws", s.serveChatWebsocket)

		log.Println("chat API listen at " + chatPath)
		log.Println("chat API accepts websocket connection at " + wsRoute.Path)
	}

	// serve static content
	e.Static("/", "")

	// start server
	serverURL := s.conf.HTTP
	log.Println("server listen at " + serverURL)
	err := e.Start(serverURL)
	e.Logger.Error(err)
	return err
}

func (s *Server) Shutdown(ctx context.Context) error {
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
	return NewServer(repos, conf).ListenAndServe()
}
