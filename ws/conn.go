package ws

import (
	"context"
	"errors"
	"io"
	"sync"

	"github.com/shirasudon/go-chat/chat/action"
	"github.com/shirasudon/go-chat/domain"
	"github.com/shirasudon/go-chat/domain/event"

	"golang.org/x/net/websocket"
)

// ActionJSON is a data-transfer-object
// which is sent by json the client connection.
type ActionJSON struct {
	ActionName action.Action     `json:"action"`
	Data       action.AnyMessage `json:"data"`
}

// Conn is end-point for reading/writing messages from/to websocket.
// One Conn corresponds to one browser-side client.
type Conn struct {
	userID   uint64
	userName string

	conn *websocket.Conn

	mu     *sync.Mutex
	closed bool          // under mu
	done   chan struct{} // done is managed by closed.

	messages chan interface{}

	onActionMessage func(*Conn, action.ActionMessage)
	onClosed        func(*Conn)
	onError         func(*Conn, error)
}

func NewConn(conn *websocket.Conn, user domain.User) *Conn {
	return &Conn{
		userID:   user.ID,
		userName: user.Name,
		conn:     conn,
		mu:       new(sync.Mutex),
		closed:   false,
		messages: make(chan interface{}, 1),
		done:     make(chan struct{}, 1),
	}
}

func (c *Conn) UserID() uint64 {
	return c.userID
}

// set callback function to handle the event for a message is received.
// the callback function may be called asynchronously.
func (c *Conn) OnActionMessage(f func(*Conn, action.ActionMessage)) {
	c.onActionMessage = f
}

// set callback function to handle the event for the connection is closed .
// the callback function may be called asynchronously.
func (c *Conn) OnClosed(f func(*Conn)) {
	c.onClosed = f
}

// set callback function to handle the event for the connection gets error.
// the callback function may be called asynchronously.
func (c *Conn) OnError(f func(*Conn, error)) {
	c.onError = f
}

// Send ActionMessage to browser-side client.
// message is ignored when Conn is closed.
func (c *Conn) Send(m event.Event) {
	select {
	case c.messages <- m:
	case <-c.done:
	}
}

var ErrAlreadyClosed = errors.New("already closed")

// Close stops Listen() immediately.
// closed Conn never listen any message.
// it returns ErrAlreadyClosed when the Conn is
// already closed otherwise nil.
func (c *Conn) Close() error {
	c.mu.Lock()
	if c.closed {
		c.mu.Unlock()
		return ErrAlreadyClosed
	}

	c.closed = true
	close(c.done)
	c.mu.Unlock() // to avoid dead lock, Unlock before OnClosed.

	if c.onClosed != nil {
		c.onClosed(c)
	}
	return nil
}

// Listen starts handling reading/writing websocket.
// it blocks until websocket is closed or context is done.
//
// when Listen() ends, Conn is closed.
func (c *Conn) Listen(ctx context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	defer func() {
		c.Close()
		cancel()
	}()

	// signal of receivePump is done
	receiveDoneCh := make(chan struct{}, 1)
	go func() {
		defer close(receiveDoneCh)
		c.receivePump(ctx)
	}()
	c.sendPump(ctx, receiveDoneCh)
}

func (c *Conn) sendPump(ctx context.Context, receiveDoneCh chan struct{}) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-c.done:
			return
		case <-receiveDoneCh:
			return
		case m := <-c.messages:
			if err := websocket.JSON.Send(c.conn, m); err != nil {
				// io.EOF means connection is closed
				if err == io.EOF {
					return
				}
				if c.onError != nil {
					c.onError(c, err)
				}
			}
		}
	}
}

func (c *Conn) receivePump(ctx context.Context) {
	// receivePump run on other goroutine.
	// done channel is not listened.
	for {
		select {
		case <-ctx.Done():
			return
		default:
			message, err := c.receiveActionJSON()
			if err != nil {
				if err == io.EOF {
					return
				}
				// return error message to client
				c.Send(event.ErrorRaised{Message: err.Error()})
				continue
			}
			// Receive success, handling received message
			c.handleActionJSON(message)
		}
	}
}

// return fatal error, such as io.EOF with connection closed,
// otherwise handle itself.
func (c *Conn) receiveActionJSON() (*ActionJSON, error) {
	var message ActionJSON
	if err := websocket.JSON.Receive(c.conn, &message); err != nil {
		// io.EOF means connection is closed
		if err == io.EOF {
			return nil, err
		}

		// actual error is handled by server.
		if c.onError != nil {
			c.onError(c, err)
		}
		return nil, errors.New("JSON structure must be a HashMap type")
	}

	// validate existance of data.
	if message.Data == nil {
		err := errors.New("JSON structure must have data field as HashMap type")
		// actual error is handled by server.
		if c.onError != nil {
			c.onError(c, err)
		}
		return nil, err
	}
	return &message, nil
}

func (c *Conn) handleActionJSON(m *ActionJSON) {
	data := m.Data
	if data == nil {
		data = make(action.AnyMessage)
	}
	data.SetString(action.KeyAction, string(m.ActionName))
	data.SetNumber(action.KeySenderID, float64(c.userID))

	actionMsg, err := action.ConvertAnyMessage(data)
	if err != nil {
		if c.onError != nil {
			c.onError(c, err)
		}
		c.Send(event.ErrorRaised{Message: err.Error()})
		return
	}
	if c.onActionMessage != nil {
		c.onActionMessage(c, actionMsg)
	}
}
