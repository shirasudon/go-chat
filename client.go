package chat

import (
	"context"
	"errors"
	"io"

	"github.com/mzki/chat/entity"

	"golang.org/x/net/websocket"
)

// Client is end-point for reading/writing messages from/to websocket.
// One Client corresponds to one browser-side client.
type Client struct {
	userID   uint
	userName string

	conn *websocket.Conn

	messages chan interface{}

	onAnyMessage  func(*Client, interface{})
	onChatMessage func(*Client, ChatMessage)
	onClosed      func(*Client)
	onError       func(*Client, error)
}

func NewClient(conn *websocket.Conn, user entity.User) *Client {
	return &Client{
		userID:   user.ID,
		userName: user.Name,
		conn:     conn,
		messages: make(chan interface{}, 1),
	}
}

// Listen starts handing reading/writing websocket.
// it blocks until websocket is closed or context is done.
func (c *Client) Listen(ctx context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	go c.sendPump(ctx)
	c.receivePump(ctx)
}

func (c *Client) sendPump(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
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

func (c *Client) receivePump(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
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

func (c *Client) handleClientMessage() error {
	var message AnyMessage
	if err := websocket.JSON.Receive(c.conn, &message); err != nil {
		return err
	}

	if action := message.Action(); action != ActionEmpty {
		return c.handleArbitraryValue(message, Action(action))
	}
	return errors.New("got json without action field")
}

func (c *Client) handleArbitraryValue(m AnyMessage, action Action) error {
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
func (c *Client) Send(v interface{}) {
	c.messages <- v
}
