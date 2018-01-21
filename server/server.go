package server

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"path"
	"sort"
	"strings"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"

	"github.com/shirasudon/go-chat/chat"
	"github.com/shirasudon/go-chat/chat/action"
	"github.com/shirasudon/go-chat/domain/event"
	"github.com/shirasudon/go-chat/ws"
)

// it represents server which can accepts chat room and its clients.
type Server struct {
	echo *echo.Echo

	wsServer     *ws.Server
	loginHandler *LoginHandler
	restHandler  *RESTHandler

	chatHub chat.Hub

	conf Config
}

// it returns new constructed server with config.
// nil config is ok and use DefaultConfig insteadly.
func NewServer(chatCmd chat.CommandService, chatQuery chat.QueryService, chatHub chat.Hub, login chat.LoginService, conf *Config) *Server {
	if conf == nil {
		conf = &DefaultConfig
	}

	e := echo.New()
	e.HideBanner = true

	s := &Server{
		echo:         e,
		loginHandler: NewLoginHandler(login),
		restHandler:  NewRESTHandler(chatCmd, chatQuery),
		chatHub:      chatHub,
		conf:         *conf,
	}
	s.wsServer = ws.NewServerFunc(s.handleWsConn)

	// initilize router
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

	chatPath := path.Join(s.conf.ChatAPIPrefix, "/chat")
	chatGroup := e.Group(chatPath, s.loginHandler.Filter())

	// set restHandler
	chatGroup.POST("/rooms", s.restHandler.CreateRoom).
		Name = "chat.createRoom"
	chatGroup.DELETE("/rooms/:room_id", s.restHandler.DeleteRoom).
		Name = "chat.deleteRoom"
	chatGroup.GET("/rooms/:room_id", s.restHandler.GetRoomInfo).
		Name = "chat.getRoomInfo"
	chatGroup.POST("/rooms/:room_id/members", s.restHandler.AddRoomMember).
		Name = "chat.addRoomMember"
	chatGroup.DELETE("/rooms/:room_id/members", s.restHandler.RemoveRoomMember).
		Name = "chat.removeRoomMember"

	chatGroup.GET("/users/:user_id", s.restHandler.GetUserInfo).
		Name = "chat.getUserInfo"

	chatGroup.POST("/rooms/:room_id/messages", s.restHandler.PostRoomMessage).
		Name = "chat.postRoomMessage"
	chatGroup.GET("/rooms/:room_id/messages", s.restHandler.GetRoomMessages).
		Name = "chat.getRoomMessages"
	chatGroup.POST("/rooms/:room_id/messages/read", s.restHandler.ReadRoomMessages).
		Name = "chat.readRoomMessages"
	chatGroup.GET("/rooms/:room_id/messages/unread", s.restHandler.GetUnreadRoomMessages).
		Name = "chat.getUnreadRoomMessages"

	// set websocket handler
	chatGroup.GET("/ws", s.serveChatWebsocket).
		Name = "chat.connentWebsocket"

	// serve static content
	if s.conf.EnableServeStaticFile {
		route := path.Join(s.conf.StaticHandlerPrefix, "/")
		e.Static(route, s.conf.StaticFileDir).Name = "staticContents"
	}
	return s
}

func (s *Server) handleWsConn(conn *ws.Conn) {
	log.Println("Server.acceptWSConn: ")
	defer conn.Close()

	var ctx = conn.Request().Context()

	conn.OnActionMessage(func(conn *ws.Conn, m action.ActionMessage) {
		s.chatHub.Send(conn, m)
	})
	conn.OnError(func(conn *ws.Conn, err error) {
		conn.Send(event.ErrorRaised{Message: err.Error()})
	})
	conn.OnClosed(func(conn *ws.Conn) {
		s.chatHub.Disconnect(conn)
	})

	err := s.chatHub.Connect(ctx, conn)
	if err != nil {
		conn.Send(event.ErrorRaised{Message: err.Error()})
		log.Printf("websocket connect error: %v\n", err)
		return
	}

	// blocking to avoid connection closed
	conn.Listen(ctx)
}

func (s *Server) serveChatWebsocket(c echo.Context) error {
	// LoggedInUserID is valid at middleware layer, loginHandler.Filter.
	userID, ok := LoggedInUserID(c)
	if !ok {
		return errors.New("needs logged in, but access without logged in state")
	}

	s.wsServer.ServeHTTPWithUserID(c.Response(), c.Request(), userID)
	return nil
}

// Handler returns http.Handler interface in the server.
func (s *Server) Handler() http.Handler {
	return s.echo
}

// it starts server process.
// it blocks until process occurs any error and
// return the error.
func (s *Server) ListenAndServe() error {
	if err := s.conf.Validate(); err != nil {
		return fmt.Errorf("server: config erorr: %v", err)
	}

	e := s.echo

	// show registered URLs
	if s.conf.ShowRoutes {
		routes := e.Routes()
		sort.Slice(routes, func(i, j int) bool {
			ri, rj := routes[i], routes[j]
			return len(ri.Path) < len(rj.Path)
		})
		for _, url := range routes {
			// built-in routes are ignored
			if strings.Contains(url.Name, "github.com/labstack/echo") {
				continue
			}
			log.Printf("%8s : %-35s (%v)\n", url.Method, url.Path, url.Name)
		}
	}

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
func ListenAndServe(chatCmd chat.CommandService, chatQuery chat.QueryService, chatHub chat.Hub, login chat.LoginService, conf *Config) error {
	s := NewServer(chatCmd, chatQuery, chatHub, login, conf)
	defer s.Shutdown(context.Background())
	return s.ListenAndServe()
}
