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
	connects          chan *Conn
	disconnects       chan *Conn
	messages          chan ActionMessage
	enterRoomRequests chan enterRoomRequest
	errors            chan error

	// TODO RoomRepository to accept that correct user enters a room.
	rooms   *RoomManager
	clients *ClientManager
}

func NewInitialRoom(repos entity.Repositories) *InitialRoom {
	// TODO set repos to field
	return &InitialRoom{
		connects:          make(chan *Conn, 1),
		disconnects:       make(chan *Conn, 1),
		messages:          make(chan ActionMessage, 1),
		enterRoomRequests: make(chan enterRoomRequest, 1),
		errors:            make(chan error, 1),
		rooms:             NewRoomManager( /*rRepo*/ ),
		clients:           NewClientManager(),
	}
}

func (iroom *InitialRoom) Listen(ctx context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	for {
		select {
		case c := <-iroom.connects:
			c.onClosed = func(closedC *Conn) {
				iroom.disconnects <- closedC
			}
			c.onError = func(conn *Conn, err error) {
				iroom.errors <- err
			}
			c.onActionMessage = func(gotC *Conn, m ActionMessage) {
				// upgrade enter room message to enter room request
				switch m.Action() {
				case ActionEnterRoom:
					iroom.enterRoomRequests <- enterRoomRequest{m.(EnterRoom), gotC}
				default:
					iroom.messages <- m
				}
			}
			iroom.clients.connectClient(c)

		case c := <-iroom.disconnects:
			c.onActionMessage = nil
			c.onError = nil
			c.onClosed = nil

			roomID := iroom.clients.roomIDFromConn(c)
			iroom.rooms.DisconnectClient(roomID, c)
			iroom.clients.disconnectClient(c)

		case m := <-iroom.messages:
			iroom.handleMessage(ctx, m)

		case req := <-iroom.enterRoomRequests:
			iroom.enterRoom(ctx, req)

		case err := <-iroom.errors:
			// TODO error handling
			log.Printf("Error: %v\n", err)

		case <-ctx.Done():
			return
		}
	}
}

type enterRoomRequest struct {
	EnterRoom
	Conn *Conn
}

// Connect new websocket client to room.
// it blocks until context is done.
func (iroom *InitialRoom) Connect(ctx context.Context, conn *websocket.Conn, u entity.User) {
	c := NewConn(conn, u)
	iroom.connects <- c
	c.Listen(ctx)
}

func (iroom *InitialRoom) handleMessage(ctx context.Context, m ActionMessage) {
	switch r := m.(type) {
	case ToRoomMessage:
		iroom.rooms.Send(r)
	}
}

func (iroom *InitialRoom) enterRoom(ctx context.Context, req enterRoomRequest) {
	if err := iroom.clients.validateClientHasRoom(req.Conn, req.SenderID, req.RoomID); err != nil {
		sendError(req.Conn, err)
		return
	}
	iroom.rooms.EnterRoom(ctx, req.CurrentRoomID, req.RoomID, req.Conn)
}

func sendError(c *Conn, err error) {
	log.Println(err)
	go func() { c.Send(NewErrorMessage(err)) }()
}
