package chat

import (
	"context"
	"log"
)

// Room is a chat room which contains multiple
// users and manages the communication between those users.
type Room struct {
	name string

	// event channels
	joins    chan *Client
	leaves   chan *Client
	messages chan ChatMessage
	errors   chan error

	onClosed func(*Room)

	repo    MessageRepository
	clients map[*Client]bool
}

func NewRoom(name string) *Room {
	return &Room{
		name:     name,
		joins:    make(chan *Client, 1),
		leaves:   make(chan *Client, 1),
		messages: make(chan ChatMessage, 1),
		errors:   make(chan error, 1),

		repo:    NewMessageRepositoryStub(),
		clients: make(map[*Client]bool, 4),
	}
}

func (room *Room) Listen(ctx context.Context) {
	log.Printf("Room(%s).Listen", room.name)
	ctx, cancel := context.WithCancel(ctx)
	defer func() { // finalize actions.
		if room.onClosed != nil {
			room.onClosed(room)
		}
		cancel()
	}()

	for {
		select {
		case c := <-room.joins:
			log.Printf("Room(%s).Join", room.name)
			room.join(c)
		case c := <-room.leaves:
			log.Printf("Room(%s).Leave", room.name)
			room.leave(c)
		case m := <-room.messages:
			log.Printf("Room(%s).Message", room.name)
			room.broadcast(m)
		case err := <-room.errors:
			log.Printf("Room(%s).Error", room.name)
			// TODO err handling
			log.Printf("Error Room(%s): %v", room.name, err)
		case <-ctx.Done():
			log.Printf("Room(%s).ContextDone", room.name)
			return
		}
		// check whether room is exist?
		if len(room.clients) == 0 {
			log.Printf("Room(%s).NoClientEnd", room.name)
			return
		}
	}
}

func (room *Room) join(c *Client) {
	// TODO how over wrapped client is handled?
	room.clients[c] = true

	c.onChatMessage = func(c *Client, m ChatMessage) {
		room.messages <- m
	}
	c.onError = func(c *Client, err error) {
		room.errors <- err
	}
	c.onClosed = func(c *Client) {
		room.leaves <- c
	}

	// send past messages to new client
	msgs, err := room.repo.Get()
	if err != nil {
		room.errors <- err
		return
	}
	for _, m := range msgs {
		c.Send(m)
	}
}

func (room *Room) leave(c *Client) {
	if _, exist := room.clients[c]; exist {
		c.onChatMessage = nil
		c.onError = nil
		c.onClosed = nil
		delete(room.clients, c)
	}
}

func (room *Room) broadcast(m ChatMessage) {
	if err := room.repo.Put(m); err != nil {
		room.errors <- err
	}
	for c, _ := range room.clients {
		c.Send(m)
	}
}

func (room *Room) Send(m ChatMessage) {
	room.messages <- m
}

func (room *Room) Join(c *Client) {
	room.joins <- c
}
