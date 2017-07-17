package chat

import (
	"context"
	"io"

	"golang.org/x/net/websocket"
)

type Client struct {
	conn *websocket.Conn

	messages chan Message

	onMessage func(*Client, Message)
	onClosed  func(*Client)
	onError   func(*Client, error)

	id   int64
	name string
}

func NewClient(conn *websocket.Conn) *Client {
	return &Client{
		conn:     conn,
		messages: make(chan Message, 1),
	}
}

func (c *Client) setID(id int64) {
	c.id = id
}

func (c *Client) Listen(ctx context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	go c.sendPump(ctx)
	c.receivePump(ctx)
}

func (c *Client) sendPump(ctx context.Context) {
	for {
		select {
		case m := <-c.messages:
			err := websocket.JSON.Send(c.conn, &m)
			if err != nil {
				// closing connection is done by readPump, so no-op.
				return
			}
		case <-ctx.Done():
			return
		}
	}
}

func (c *Client) receivePump(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			var msg Message
			err := websocket.JSON.Receive(c.conn, &msg)
			switch err {
			case nil:
				// Receive success
				if c.onMessage != nil {
					c.onMessage(c, msg)
				}
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

func (c *Client) Send(m Message) {
	c.messages <- m
}
