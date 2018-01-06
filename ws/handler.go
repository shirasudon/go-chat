package ws

import (
	"context"
	"errors"
	"log"
	"net/http"

	"github.com/shirasudon/go-chat/domain/event"
	"golang.org/x/net/websocket"
)

// Handler handles websocket Conn in this package.
type Handler func(*Conn)

// Server serves Conn, wrapper for websocket Connetion, for each HTTP request.
// It implements http.Handler interface.
type Server struct {
	server *websocket.Server

	// Handler for the Conn type in this package.
	Handler Handler
}

// NewServer creates the server which serves websocket Connection and
// providing customizable handler for that connection.
func NewServer(handler Handler) *Server {
	if handler == nil {
		panic("nil handler")
	}
	s := &Server{Handler: handler}
	s.server = &websocket.Server{Handler: s.wsHandler}
	return s
}

// NewServerFunc is wrapper function for the NewServer so that
// given function need not to cast to Handler type.
func NewServerFunc(handler func(*Conn)) *Server {
	return NewServer(Handler(handler))
}

func (s *Server) wsHandler(wsConn *websocket.Conn) {
	defer wsConn.Close()

	userID, err := getConnectUserID(wsConn.Request().Context())
	if err != nil {
		websocket.JSON.Send(wsConn, event.ErrorRaised{Message: "invalid state"})
		// TODO logging by external logger
		log.Printf("websocket server can not handling this connection, error: %v\n", err)
		return // to close connection.
	}

	c := NewConn(wsConn, userID)
	s.Handler(c)
}

// ServeHTTPWithUserID is similar with the http.Handler except that
// it requires userID to specify the which user connects.
func (s *Server) ServeHTTPWithUserID(w http.ResponseWriter, req *http.Request, userID uint64) {
	newCtx := setConnectUserID(req.Context(), userID)
	s.server.ServeHTTP(w, req.WithContext(newCtx))
}

const ctxKeyConnectUserID = "_ws_connect_user_id"

func setConnectUserID(ctx context.Context, userID uint64) context.Context {
	return context.WithValue(ctx, ctxKeyConnectUserID, userID)
}

func getConnectUserID(ctx context.Context) (uint64, error) {
	userID, ok := ctx.Value(ctxKeyConnectUserID).(uint64)
	if !ok {
		return 0, errors.New("requested websocket connection has no user ID")
	}
	return userID, nil
}
