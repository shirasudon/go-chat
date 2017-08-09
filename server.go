package chat

import (
	"context"
	"errors"
	"log"

	"golang.org/x/net/websocket"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/shirasudon/go-chat/entity"
	"github.com/shirasudon/go-chat/model"
)

// it represents server which can accepts chat room and its clients.
type Server struct {
	websocketServer *websocket.Server
	loginHandler    *LoginHandler
	initialRoom     *model.InitialRoom

	ctx context.Context

	repos entity.Repositories

	conf Config
}

// it returns new constructed server with config.
// nil config is ok and use DefaultConfig insteadly.
func NewServer(repos entity.Repositories, conf *Config) *Server {
	if conf == nil {
		conf = &DefaultConfig
	}

	s := &Server{
		loginHandler: NewLoginHandler(repos.Users()),
		initialRoom:  model.NewInitialRoom(repos),
		ctx:          context.Background(),
		repos:        repos,
		conf:         *conf,
	}
	s.websocketServer = &websocket.Server{} // TODO it is needed?
	return s
}

func (s *Server) serveChatWebsocket(c echo.Context) error {
	userID, ok := c.Get(KeyLoggedInUserID).(uint64)
	if !ok {
		return errors.New("needs logged in, but access without logged in state")
	}
	user, err := s.repos.Users().Find(userID)
	if err != nil {
		return err
	}
	websocket.Handler(func(ws *websocket.Conn) {
		log.Println("Server.acceptWSConn: ")
		defer ws.Close()

		// blocking to avoid connection closed
		s.initialRoom.Connect(s.ctx, ws, user)
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
	s.ctx = ctx // override context to propagate done signal from the server.
	defer cancel()

	// initilize router
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// set login handler
	e.Use(s.loginHandler.Middleware())
	e.POST("/login", s.loginHandler.Login)
	e.GET("/login", s.loginHandler.GetLoginState)
	e.POST("/logout", s.loginHandler.Logout)

	// start chat proces
	s.initialRoom.Listen(ctx)

	// set websocket handler
	g := e.Group("/ws", s.loginHandler.Filter())
	g.GET(s.conf.WebSocketPath, s.serveChatWebsocket)

	// serve static content
	e.Static("/", "")

	// start server
	serverURL := s.conf.HTTP
	log.Println("WebSocket server listen at " + serverURL + s.conf.WebSocketPath)
	err := e.Start(serverURL)
	e.Logger.Error(err)
	return err
}

// It starts server process using default server with
// user config.
// A nil config is OK and use DefaultConfig insteadly.
// It blocks until the process occurs any error and
// return the error.
func ListenAndServe(repos entity.Repositories, conf *Config) error {
	if conf == nil {
		conf = &DefaultConfig
	}
	return NewServer(repos, conf).ListenAndServe()
}
