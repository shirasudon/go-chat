package chat

import (
	"context"
	"log"
	"net/http"
	"path"
	"strings"
	"sync"

	"golang.org/x/net/websocket"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/mzki/chat/entity"
	"github.com/mzki/chat/model"
)

// it represents server which can accepts chat room and its clients.
type Server struct {
	websocketServer *websocket.Server
	loginHandler    *LoginHandler

	ctx context.Context

	repos entity.Repositories

	mutex *sync.RWMutex
	rooms map[string]*model.Room
	conf  Config
}

// it returns new constructed server with config.
// nil config is ok and use DefaultConfig insteadly.
func NewServer(repos entity.Repositories, conf *Config) *Server {
	if conf == nil {
		conf = &DefaultConfig
	}

	s := &Server{
		loginHandler: NewLoginHandler(),
		ctx:          context.Background(),
		repos:        repos,
		mutex:        new(sync.RWMutex),
		rooms:        make(map[string]*model.Room, 4),
		conf:         *conf,
	}
	s.websocketServer = &websocket.Server{Handler: websocket.Handler(s.acceptRoom)}
	return s
}

func (s *Server) acceptRoom(ws *websocket.Conn) {
	defer ws.Close()

	log.Println("Server.acceptRoom: " + ws.Request().URL.String())
	room_id := path.Base(ws.Request().URL.Path)

	s.mutex.Lock()
	room, exist := s.rooms[room_id]
	if !exist {
		room = model.NewRoom(room_id)
		room.OnClosed = s.doneRoom
		go room.Listen(s.ctx)
		s.rooms[room_id] = room
	}
	s.mutex.Unlock()

	c := model.NewClient(ws, entity.User{}) // TODO session's user
	room.Join(c)
	c.Listen(s.ctx) // blocking to avoid connection closed
}

func (s *Server) doneRoom(r *model.Room) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	r.OnClosed = nil
	delete(s.rooms, r.Name())
}

// it starts server process.
// it blocks until process occurs any error and
// return the error.
func (s *Server) ListenAndServe() error {
	if err := s.conf.validate(); err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(context.Background())
	s.ctx = ctx // overwrite context to propagate cancel siganl to others.
	defer cancel()

	// initilize router
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.Static("/", "")
	e.GET(s.conf.WebSocketPath+"*", func(c echo.Context) error {
		s.routingRoom(c.Response(), c.Request())
		return nil
	})

	// start server
	serverURL := s.conf.HTTP
	log.Println("WebSocket server listen at " + serverURL + s.conf.WebSocketPath)
	err := e.Start(serverURL)
	e.Logger.Error(err)
	return err
}

func (s *Server) routingRoom(w http.ResponseWriter, r *http.Request) {
	log.Println("routingRoom: " + r.URL.String())

	room_id := strings.TrimPrefix(r.URL.Path, s.conf.WebSocketPath)
	if len(room_id) > 0 && !strings.Contains(room_id, "/") {
		// serve websocket
		s.websocketServer.ServeHTTP(w, r)
	} else {
		http.Error(w, "can not match any rooms", http.StatusBadRequest)
	}
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
