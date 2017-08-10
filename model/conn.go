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
				if err != io.EOF && c.onError != nil {
					c.onError(c, err)
				}
				return
			}
		}
	}
}

func (c *Conn) receivePump(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		// receivePump run on other goroutine.
		// done channel is not listened.
		default:
			err := c.handleClientMessage()
			switch err {
			case nil:
				// Receive success, no-op
			default:
				// io.EOF means connection is closed
				if err != io.EOF && c.onError != nil {
					c.onError(c, err)
				}
				return // to receivePump()
			}
		}
	}
}

func (c *Conn) handleClientMessage() error {
	var message AnyMessage
	if err := websocket.JSON.Receive(c.conn, &message); err != nil {
		return err
	}
	return c.handleAnyMessae(message)
}

func (c *Conn) handleAnyMessae(m AnyMessage) error {
	var actionMessage ActionMessage
	switch action := m.Action(); action {
	case ActionChatMessage:
		actionMessage = ParseChatMessage(m, action)

	case ActionReadMessage:
		actionMessage = ParseReadMessage(m, action)

	case ActionTypeStart:
		typing := ParseTypeStart(m, action)
		typing.SenderID = c.userID
		typing.SenderName = c.userName
		// restore as actionMessage
		actionMessage = typing

	case ActionTypeEnd:
		typing := ParseTypeEnd(m, action)
		typing.SenderID = c.userID
		typing.SenderName = c.userName
		// restore as actionMessage
		actionMessage = typing

	case ActionEmpty:
		return errors.New("got json without action field")
	default:
		return errors.New("handleAnyMessae: unknown action: " + string(action))
	}

	// send parsed ActionMessage to onActionMessage handler.
	if actionMessage != nil && c.onActionMessage != nil {
		c.onActionMessage(c, actionMessage)
	}
	return nil
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
