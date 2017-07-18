package chat

import (
	"context"
	"log"
	"net/http"
	"path"
	"strings"
	"sync"

	"golang.org/x/net/websocket"
)

// it represents server which can accepts chat room and its clients.
type Server struct {
	server          *http.Server
	websocketServer *websocket.Server

	ctx context.Context

	mutex *sync.RWMutex
	rooms map[string]*Room
	conf  Config
}

// it returns new constructed server with config.
// nil config is ok and use DefaultConfig insteadly.
func NewServer(conf *Config) *Server {
	if conf == nil {
		conf = &DefaultConfig
	}

	s := &Server{
		server: &http.Server{},
		ctx:    context.Background(),
		mutex:  new(sync.RWMutex),
		rooms:  make(map[string]*Room, 4),
		conf:   *conf,
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
		room = NewRoom(room_id)
		room.onClosed = s.doneRoom
		go room.Listen(s.ctx)
		s.rooms[room_id] = room
	}
	s.mutex.Unlock()

	c := NewClient(ws)
	room.Join(c)
	c.Listen(s.ctx) // blocking to avoid connection closed
}

func (s *Server) doneRoom(r *Room) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	r.onClosed = nil
	delete(s.rooms, r.name)
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

	serverURL := s.conf.HTTP

	handler := http.NewServeMux()
	handler.Handle("/", http.FileServer(http.Dir("./")))
	handler.HandleFunc(s.conf.WebSocketPath, s.routingRoom)
	log.Println("WebSocket server listen at " + serverURL + s.conf.WebSocketPath)

	s.server.Addr = serverURL
	s.server.Handler = handler
	return s.server.ListenAndServe()
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
func ListenAndServe(conf *Config) error {
	if conf == nil {
		conf = &DefaultConfig
	}
	return NewServer(conf).ListenAndServe()
}
