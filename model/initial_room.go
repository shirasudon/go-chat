package model

import (
	"context"
	"fmt"
	"log"

	"github.com/shirasudon/go-chat/entity"
	"golang.org/x/net/websocket"
)

// connection specified infomation.
type connectionInfo struct {
	currentRoomID uint64
}

// activeClient is a one active user which has several
// websocket connections, rooms, and friends.
//
// activeClient has multiple websocket conncetion,
// because one user has many websocket conncetions such as from PC,
// mobile device, and so on.
type activeClient struct {
	conns   map[*Conn]connectionInfo
	rooms   map[uint64]bool // has rooms managed by room id
	friends map[uint64]bool // has friends managed by user id
}

func newActiveClient(c *Conn) *activeClient {
	return &activeClient{
		conns:   map[*Conn]connectionInfo{c: connectionInfo{}},
		rooms:   make(map[uint64]bool),
		friends: make(map[uint64]bool),
	}
}

func (ac *activeClient) Send(m ActionMessage) {
	for c, _ := range ac.conns {
		c.Send(m)
	}
}

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
	roomManager *RoomManager
	clients     map[uint64]*activeClient
}

func NewInitialRoom( /*rRepo RoomRepository*/ ) *InitialRoom {
	return &InitialRoom{
		connects:    make(chan *Conn, 1),
		disconnects: make(chan *Conn, 1),
		messages:    make(chan ActionMessage, 1),
		roomManager: NewRoomManager( /*rRepo*/ ),
		clients:     make(map[uint64]*activeClient),
	}
}

func (iroom *InitialRoom) Listen(ctx context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	for {
		select {
		case c := <-iroom.connects:
			iroom.connectClient(c)
		case c := <-iroom.disconnects:
			iroom.disconnectClient(c)
		case m := <-iroom.messages:
			iroom.handleMessage(ctx, m)
		case <-ctx.Done():
			return
		}
	}
}

func (iroom *InitialRoom) connectClient(c *Conn) {
	// add conncetion to active user when the user is already connected from anywhere.
	if activeC, ok := iroom.clients[c.userID]; ok {
		activeC.conns[c] = connectionInfo{}
		return
	}

	// create new active client because the connection is newly.
	// and broadcasts user connect event to all active friends
	activeC := newActiveClient(c)
	// TODO set friends and rooms to active user.
	iroom.clients[c.userID] = activeC
	uc := UserConnect{} // TODO implement UserConncet.
	iroom.broadcastsFriends(activeC, uc)
}

func (iroom *InitialRoom) broadcastsFriends(activeC *activeClient, m ActionMessage) {
	for friendID, _ := range activeC.friends {
		if activeFriend, ok := iroom.clients[friendID]; ok {
			activeFriend.Send(m)
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

func (iroom *InitialRoom) disconnectClient(c *Conn) {
	activeC, ok := iroom.clients[c.userID]
	if !ok {
		return
	}
	connInfo, ok := activeC.conns[c]
	if !ok {
		return
	}
	iroom.roomManager.DisconnectClient(connInfo.currentRoomID, c)

	delete(activeC.conns, c)
	if len(activeC.conns) == 0 {
		delete(iroom.clients, c.userID)
		// udc := UserDisconnect{} // TODO implement UserDisconncet.
		// iroom.broadcastsFriends(activeC, udc)
	}
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
	if err := iroom.validateClientHasRoom(req.Client, req.SenderID, req.RoomID); err != nil {
		sendError(req.Client, err)
		return
	}

	iroom.roomManager.EnterRoom(ctx, req.CurrentRoomID, req.RoomID, req.Client)
}

func (iroom *InitialRoom) validateClientHasRoom(conn *Conn, userID, roomID uint64) error {
	if !iroom.connectionExist(userID, conn) {
		return fmt.Errorf("request user id(%d) is not found", userID)
	}
	// the client is validated at above.
	if _, ok := iroom.clients[userID].rooms[roomID]; !ok {
		return fmt.Errorf("user(%d) does not have request room id(%d) ", userID, roomID)
	}
	return nil
}

// check whether active client with the websocket connection exists?
func (iroom *InitialRoom) connectionExist(userID uint64, conn *Conn) bool {
	if activeC, ok := iroom.clients[userID]; ok {
		if _, ok := activeC.conns[conn]; ok {
			return true
		}
	}
	return false
}

func sendError(c *Conn, err error) {
	log.Println(err)
	go func() { c.Send(NewErrorMessage(err)) }()
}
