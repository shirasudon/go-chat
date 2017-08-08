package model

import (
	"context"
	"errors"
	"io"

	"github.com/shirasudon/go-chat/entity"

	"golang.org/x/net/websocket"
)

// Conn is end-point for reading/writing messages from/to websocket.
// One Conn corresponds to one browser-side client.
type Conn struct {
	userID   uint64
	userName string

	conn *websocket.Conn

	messages chan interface{}

	onAnyMessage  func(*Conn, interface{})
	onChatMessage func(*Conn, ChatMessage)
	onClosed      func(*Conn)
	onError       func(*Conn, error)
}

func NewConn(conn *websocket.Conn, user entity.User) *Conn {
	return &Conn{
		userID:   user.ID,
		userName: user.Name,
		conn:     conn,
		messages: make(chan interface{}, 1),
	}
}

// Listen starts handing reading/writing websocket.
// it blocks until websocket is closed or context is done.
func (c *Conn) Listen(ctx context.Context) {
	// two Pump functions are listening the closing event for each other.
	sendingDone := make(chan struct{})
	receivingDone := make(chan struct{})
	go func() {
		defer close(sendingDone)
		c.sendPump(ctx, receivingDone)
	}()
	defer close(receivingDone)
	c.receivePump(ctx, sendingDone)
}

func (c *Conn) sendPump(ctx context.Context, receivingDone chan struct{}) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-receivingDone:
			return
		case m := <-c.messages:
			err := websocket.JSON.Send(c.conn, &m)
			if err == io.EOF {
				// closing connection is done by readPump, so no-op.
				return
			}
		}
	}
}

func (c *Conn) receivePump(ctx context.Context, sendingDone chan struct{}) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-sendingDone:
			return
		default:
			err := c.handleClientMessage()
			switch err {
			case nil:
				// Receive success, no-op
			case io.EOF:
				// connection is closed
				if c.onClosed != nil {
					c.onClosed(c)
				}
				return
			default:
				if c.onError != nil {
					c.onError(c, err)
				}
			}
		}
	}
}

func (c *Conn) handleClientMessage() error {
	var message AnyMessage
	if err := websocket.JSON.Receive(c.conn, &message); err != nil {
		return err
	}

	if action := message.Action(); action != ActionEmpty {
		return c.handleArbitraryValue(message, Action(action))
	}
	return errors.New("got json without action field")
}

func (c *Conn) handleArbitraryValue(m AnyMessage, action Action) error {
	switch action {
	case ActionChatMessage:
		message := ParseChatMessage(m, action)
		if c.onChatMessage != nil {
			c.onChatMessage(c, message)
		}

	case ActionReadMessage:
		message := ParseReadMessage(m, action)
		if c.onAnyMessage != nil {
			c.onAnyMessage(c, message)
		}

	case ActionTypeStart:
		typing := ParseTypeStart(m, action)
		typing.SenderID = c.userID
		typing.SenderName = c.userName
		if c.onAnyMessage != nil {
			c.onAnyMessage(c, typing)
		}

	case ActionTypeEnd:
		typing := ParseTypeEnd(m, action)
		typing.SenderID = c.userID
		typing.SenderName = c.userName
		if c.onAnyMessage != nil {
			c.onAnyMessage(c, typing)
		}

	default:
		return errors.New("handleArbitrayValue: unknown action: " + string(action))
	}
	return nil
}

// Send aribitrary value to browser-side client.
func (c *Conn) Send(v interface{}) {
	c.messages <- v
}
