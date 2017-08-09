package model

import (
	"context"
	"log"

	"github.com/shirasudon/go-chat/entity"
)

// Room is a chat room which contains multiple
// users and manages the communication between those.
type Room struct {
	id   uint64
	name string

	// event channels
	joins    chan *Conn
	leaves   chan *Conn
	messages chan ToRoomMessage
	errors   chan error
	done     chan struct{}

	repo  entity.MessageRepository
	conns map[*Conn]bool
}

func NewRoom(name string, mRepo entity.MessageRepository) *Room {
	return &Room{
		name:     name,
		joins:    make(chan *Conn, 1),
		leaves:   make(chan *Conn, 1),
		messages: make(chan ToRoomMessage, 1),
		errors:   make(chan error, 1),
		done:     make(chan struct{}, 1),

		repo:  mRepo,
		conns: make(map[*Conn]bool, 4),
	}
}

func (room *Room) Name() string { return room.name }

func (room *Room) Listen(ctx context.Context) {
	log.Printf("Room(%s).Listen", room.name)
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	for {
		select {
		case c := <-room.joins:
			room.join(c)
		case c := <-room.leaves:
			room.leave(c)
		case m := <-room.messages:
			if err := room.handleMessage(m); err != nil {
				log.Printf("message handling Room(%s): %v", room.name, err)
			}
		case err := <-room.errors:
			// TODO err handling
			log.Printf("error Room(%s): %v", room.name, err)
		case <-room.done:
			return
		case <-ctx.Done():
			return
		}
	}
}

func (room *Room) join(c *Conn) {
	room.conns[c] = true
}

func (room *Room) leave(c *Conn) {
	if _, exist := room.conns[c]; exist {
		delete(room.conns, c)
	}
}

func (room *Room) handleMessage(m ToRoomMessage) error {
	switch m := m.(type) {
	case ChatMessage:
		var err error
		if m.ID, err = room.repo.Save(entity.Message{
			Content: m.Content,
			UserID:  m.SenderID,
			RoomID:  m.RoomID,
		}); err != nil {
			return err
		}
	}

	room.broadcast(m)
	return nil
}

func (room *Room) broadcast(m ToRoomMessage) {
	if m, ok := m.(ActionMessage); ok {
		for c, _ := range room.conns {
			c.Send(m)
		}
	}
}

func (room *Room) Send(m ToRoomMessage) {
	room.messages <- m
}

func (room *Room) Join(c *Conn) {
	room.joins <- c
}

func (room *Room) Leave(c *Conn) {
	room.leaves <- c
}

func (room *Room) Done() {
	room.done <- struct{}{}
}
