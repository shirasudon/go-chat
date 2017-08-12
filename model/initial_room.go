package model

import (
	"context"
	"log"

	"github.com/shirasudon/go-chat/entity"
	"golang.org/x/net/websocket"
)

// InitialRoom is the room which have any clients created newly.
// any clients enters this room firstly, then dispatch their to requesting rooms.
// clients enter again this room after leaving the requesting rooms, then waiting for
// dispatch to next rooms they are requested.
//
// All of the rooms are children of this room. So They are managed by InitialRoom.
type InitialRoom struct {
	connects    chan *Conn
	disconnects chan *Conn
	messages    chan actionMessageRequest
	errors      chan error

	repos   entity.Repositories
	rooms   *RoomManager
	clients *ClientManager
}

// actionMessageRequest is a composit struct of
// ActionMessage and Conn sent the message.
// It is used to handle ActionMessage by InitialRoom.
type actionMessageRequest struct {
	ActionMessage
	Conn *Conn
}

func NewInitialRoom(repos entity.Repositories) *InitialRoom {
	return &InitialRoom{
		connects:    make(chan *Conn, 1),
		disconnects: make(chan *Conn, 1),
		messages:    make(chan actionMessageRequest, 1),
		errors:      make(chan error, 1),
		repos:       repos,
		rooms:       NewRoomManager(repos),
		clients:     NewClientManager(repos),
	}
}

func (iroom *InitialRoom) Listen(ctx context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	for {
		select {
		case c := <-iroom.connects:
			c.onClosed = func(conn *Conn) { iroom.disconnects <- conn }
			c.onError = func(conn *Conn, err error) { iroom.errors <- err }
			c.onActionMessage = func(conn *Conn, m ActionMessage) {
				iroom.messages <- actionMessageRequest{m, conn}
			}

			iroom.clients.connectClient(ctx, c)

		case c := <-iroom.disconnects:
			c.onActionMessage = nil
			c.onError = nil
			c.onClosed = nil

			roomID := iroom.clients.roomIDFromConn(c)
			iroom.rooms.DisconnectClient(roomID, c)
			iroom.clients.disconnectClient(c)

		case m := <-iroom.messages:
			iroom.handleMessage(ctx, m)

		case err := <-iroom.errors:
			// TODO error handling
			log.Printf("Error: %v\n", err)

		case <-ctx.Done():
			return
		}
	}
}

// Connect new websocket client to room.
// it blocks until context is done.
func (iroom *InitialRoom) Connect(ctx context.Context, conn *websocket.Conn, u entity.User) {
	c := NewConn(conn, u)
	iroom.connects <- c
	c.Listen(ctx)
}

func (iroom *InitialRoom) handleMessage(ctx context.Context, req actionMessageRequest) {
	switch m := req.ActionMessage.(type) {
	case ToRoomMessage:
		iroom.rooms.Send(m)
	case EnterRoom:
		iroom.enterRoom(ctx, m, req.Conn)
	}
}

func (iroom *InitialRoom) enterRoom(ctx context.Context, em EnterRoom, conn *Conn) {
	if err := iroom.clients.validateClientHasRoom(conn, em.SenderID, em.RoomID); err != nil {
		sendError(conn, err, em)
		return
	}
	if err := iroom.rooms.EnterRoom(ctx, em.CurrentRoomID, em.RoomID, conn); err != nil {
		sendError(conn, err, em)
	}
}

func sendError(c *Conn, err error, cause ...ActionMessage) {
	log.Println(err)
	go func() { c.Send(NewErrorMessage(err, cause...)) }()
}
