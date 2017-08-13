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
	chatHub *ChatHub
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
		chatHub:     NewChatHub(repos),
	}
}

func (iroom *InitialRoom) Listen(ctx context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	for {
		// disconnect event has priority than others
		// because disconnected client can not be received any message.
		select {
		case c := <-iroom.disconnects:
			if err := iroom.disconnectClient(ctx, c); err != nil {
				// TODO err handling
				log.Printf("Disonnect Error: %v\n", err)
			}
			continue
		case <-ctx.Done():
			return
		default:
			// fall through if no disconnect events.
		}

		select {
		case c := <-iroom.connects:
			if err := iroom.connectClient(ctx, c); err != nil {
				// TODO: error handling
				log.Printf("Connect Error: %v\n", err)
			}

		case c := <-iroom.disconnects:
			if err := iroom.disconnectClient(ctx, c); err != nil {
				// TODO err handling
				log.Printf("Disonnect Error: %v\n", err)
			}

		case req := <-iroom.messages:
			if err := iroom.chatHub.handleMessage(ctx, req); err != nil {
				// TODO err handling
				log.Printf("Message Error: %v\n", err)
			}

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
	c.onClosed = func(conn *Conn) { iroom.disconnects <- conn }
	c.onError = func(conn *Conn, err error) { iroom.errors <- err }
	c.onActionMessage = func(conn *Conn, m ActionMessage) {
		iroom.messages <- actionMessageRequest{m, conn}
	}

	iroom.connects <- c
	c.Listen(ctx)
}

func (iroom *InitialRoom) connectClient(ctx context.Context, c *Conn) error {
	return iroom.chatHub.connectClient(ctx, c)
}

func (iroom *InitialRoom) disconnectClient(ctx context.Context, c *Conn) error {
	c.onActionMessage = nil
	c.onError = nil
	c.onClosed = nil

	return iroom.chatHub.disconnectClient(ctx, c)
}

func sendError(c *Conn, err error, cause ...ActionMessage) {
	log.Println(err)
	go func() { c.Send(NewErrorMessage(err, cause...)) }()
}
