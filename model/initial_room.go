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
	messages    chan ActionMessage

	// TODO RoomRepository to accept that correct user enters a room.
	rooms   *RoomManager
	clients *ClientManager
}

func NewInitialRoom( /*rRepo RoomRepository*/ ) *InitialRoom {
	return &InitialRoom{
		connects:    make(chan *Conn, 1),
		disconnects: make(chan *Conn, 1),
		messages:    make(chan ActionMessage, 1),
		rooms:       NewRoomManager( /*rRepo*/ ),
		clients:     NewClientManager(),
	}
}

func (iroom *InitialRoom) Listen(ctx context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	for {
		select {
		case c := <-iroom.connects:
			iroom.clients.connectClient(c)

		case c := <-iroom.disconnects:
			roomID := iroom.clients.roomIDFromConn(c)
			iroom.rooms.DisconnectClient(roomID, c)
			iroom.clients.disconnectClient(c)

		case m := <-iroom.messages:
			iroom.handleMessage(ctx, m)

		case <-ctx.Done():
			return
		}
	}
}

// Connect new websocket client to room.
// it blocks until context is done.
func (iroom *InitialRoom) Connect(ctx context.Context, conn *websocket.Conn, u entity.User) {
	c := NewConn(conn, u)
	c.onClosed = func(closedC *Conn) {
		iroom.disconnects <- closedC
	}
	c.onAnyMessage = func(gotC *Conn, msg interface{}) {
		m := msg.(ActionMessage)
		switch m.Action() {
		case ActionEnterRoom:
			m = enterRoomRequest{m.(EnterRoom), gotC}
		case ActionExitRoom:
			m = exitRoomRequest{m.(ExitRoom), gotC}
		}
		iroom.messages <- m
	}
	iroom.connects <- c
	c.Listen(ctx)
}

type enterRoomRequest struct {
	EnterRoom
	Client *Conn
}

type exitRoomRequest struct {
	ExitRoom
	Client *Conn
}

func (iroom *InitialRoom) handleMessage(ctx context.Context, m ActionMessage) {
	switch m.Action() {
	case ActionCreateRoom:
	case ActionDeleteRoom:
	case ActionEnterRoom:
		iroom.enterRoom(ctx, m.(enterRoomRequest))
	}
}

func (iroom *InitialRoom) enterRoom(ctx context.Context, req enterRoomRequest) {
	if err := iroom.clients.validateClientHasRoom(req.Client, req.SenderID, req.RoomID); err != nil {
		sendError(req.Client, err)
		return
	}

	iroom.rooms.EnterRoom(ctx, req.CurrentRoomID, req.RoomID, req.Client)
}

func sendError(c *Conn, err error) {
	log.Println(err)
	go func() { c.Send(NewErrorMessage(err)) }()
}
