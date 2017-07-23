package chat

import (
	"context"
	"errors"
	"io"

	"golang.org/x/net/websocket"
)

// Client is end-point for reading/writing messages from/to websocket.
// One Client corresponds to one browser-side client.
type Client struct {
	userID int64
	conn   *websocket.Conn

	messages chan interface{}

	onChatMessage func(*Client, ChatMessage)
	onClosed      func(*Client)
	onError       func(*Client, error)
}

func NewClient(conn *websocket.Conn, userID int64) *Client {
	return &Client{
		userID:   userID,
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
	var message Message
	if err := websocket.JSON.Receive(c.conn, &message); err != nil {
		return err
	}

	action, ok := message[KeyAction].(string)
	if !ok {
		return errors.New("got json without action field")
	}
	c.handleArbitrayValue(message, Action(action))
	return nil
}

func (c *Client) handleArbitrayValue(v Message, action Action) {
	switch action {
	case ActionChatMessage:
		message := ParseChatMessage(v, action)
		if c.onChatMessage != nil {
			c.onChatMessage(c, message)
		}
	default:
		// TODO implement
	}
}

// Send aribitrary value to browser-side client.
func (c *Client) Send(v interface{}) {
	c.messages <- v
}
