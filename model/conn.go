package model

import (
	"context"
	"errors"
	"io"
	"sync"

	"github.com/shirasudon/go-chat/entity"

	"golang.org/x/net/websocket"
)

// Conn is end-point for reading/writing messages from/to websocket.
// One Conn corresponds to one browser-side client.
type Conn struct {
	userID   uint64
	userName string

	conn *websocket.Conn

	mu     *sync.RWMutex
	closed bool // under mu

	messages chan interface{}
	done     chan struct{}

	onActionMessage func(*Conn, ActionMessage)
	onClosed        func(*Conn)
	onError         func(*Conn, error)
}

func NewConn(conn *websocket.Conn, user entity.User) *Conn {
	return &Conn{
		userID:   user.ID,
		userName: user.Name,
		conn:     conn,
		mu:       new(sync.RWMutex),
		closed:   false,
		messages: make(chan interface{}, 1),
		done:     make(chan struct{}, 1),
	}
}

// Send ActionMessage to browser-side client.
// message is ignored when Conn is closed.
func (c *Conn) Send(m ActionMessage) {
	if c.isClosed() {
		return
	}
	c.messages <- m
}

// send done signal to quit Listen() for client message.
// signal is ignored when Conn is closed.
func (c *Conn) Done() {
	if c.isClosed() {
		return
	}
	c.done <- struct{}{}
}

func (c *Conn) isClosed() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.closed
}

// Listen starts handing reading/writing websocket.
// it blocks until websocket is closed, context is done or
// Done() signal is sent.
//
// when Listen() ends, Conn is closed.
func (c *Conn) Listen(ctx context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	defer func() {
		c.mu.Lock()
		c.closed = true
		c.mu.Unlock()
		cancel()
		if c.onClosed != nil {
			c.onClosed(c)
		}
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
			message, err := c.receiveAnyMessage()
			if err == io.EOF {
				return
			}
			// Receive success, handling received message
			c.handleAnyMessage(message)
		}
	}
}

// return fatal error, such as io.EOF with connection closed,
// otherwise handle itself.
func (c *Conn) receiveAnyMessage() (AnyMessage, error) {
	var message AnyMessage
	if err := websocket.JSON.Receive(c.conn, &message); err != nil {
		// io.EOF means connection is closed
		if err == io.EOF {
			return nil, err
		}

		// actual error is handled by server.
		if c.onError != nil {
			c.onError(c, err)
		}
		// and return error message to client
		c.Send(NewErrorMessage(errors.New("JSON structure must be a HashMap type")))
	}
	return message, nil
}

func (c *Conn) handleAnyMessage(m AnyMessage) {
	action, err := c.convertAnyMessage(m)
	if err != nil {
		if c.onError != nil {
			c.onError(c, err)
		}
		c.Send(NewErrorMessage(err, action))
		return
	}
	if c.onActionMessage != nil {
		c.onActionMessage(c, action)
	}
}

func (c *Conn) convertAnyMessage(m AnyMessage) (ActionMessage, error) {
	switch action := m.Action(); action {
	case ActionChatMessage:
		cm := ParseChatMessage(m, action)
		cm.SenderID = c.userID
		return cm, nil

	case ActionReadMessage:
		rm := ParseReadMessage(m, action)
		rm.SenderID = c.userID
		return rm, nil

	case ActionTypeStart:
		typing := ParseTypeStart(m, action)
		typing.SenderID = c.userID
		typing.SenderName = c.userName
		return typing, nil

	case ActionTypeEnd:
		typing := ParseTypeEnd(m, action)
		typing.SenderID = c.userID
		typing.SenderName = c.userName
		return typing, nil

	case ActionEmpty:
		return m, errors.New("JSON object must have any action field")
	default:
		return m, errors.New("unknown action: " + string(action))
	}
}
