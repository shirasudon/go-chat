// package wstest provides utility funtions for testing websocket.

package wstest

import (
	"net/http/httptest"
	"strings"

	"golang.org/x/net/websocket"
)

// create client-side conncetion for the websocket
func NewClientConn(requestPath, origin string) (*websocket.Conn, error) {
	wsURL := strings.Replace(requestPath, "http://", "ws://", 1)
	return websocket.Dial(wsURL, "", origin)
}

// NewServer returns httptest.Server which responds
// to any request by using websocket handler.
func NewServer(wshandler websocket.Handler) *httptest.Server {
	return httptest.NewServer(wshandler)
}
