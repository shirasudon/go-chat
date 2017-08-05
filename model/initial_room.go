package model

import (
	"context"
	"fmt"
	"log"

	"github.com/shirasudon/go-chat/entity"
	"golang.org/x/net/websocket"
)

// activeClient is a one active user which has several
// websocket connections, rooms, and friends.
//
// activeClient manages multiple clients for each websocket conncetion,
// because one user has many websocket conncetions such as from PC,
// mobile device, and so on.
type activeClient struct {
	conns   map[*Client]bool
	rooms   map[uint64]bool // has rooms managed by room id
	friends map[uint64]bool // has friends managed by user id
}

func newActiveClient(c *Client) *activeClient {
	return &activeClient{
		conns:   map[*Client]bool{c: true},
		rooms:   make(map[uint64]bool),
		friends: make(map[uint64]bool),
	}
}

func (ac *activeClient) Send(m ActionMessage) {
	for f, _ := range ac.friends {
		f.Send(m)
	}
}

// activeRoom is a wrapper for Room which has
// a number of active chat members.
type activeRoom struct {
	room    *Room
	nMenber int
}

func newActiveRoom(r *Room) *activeRoom {
	return &activeRoom{
		room:    room,
		nMenber: 0,
	}
}

type enterRoomRequest struct {
	EnterRoom
	Client *Client
}

type exitRoomRequest struct {
	ExitRoom
	Client *Client
}

// InitialRoom is the room which have any clients created newly.
// any clients enters this room firstly, then dispatch their to requesting rooms.
// clients enter again this room after leaving the requesting rooms, then waiting for
// dispatch to next rooms they are requested.
//
// All of the rooms are children of this room. So They are managed by InitialRoom.
type InitialRoom struct {
	connects    chan *Client
	disconnects chan *Client
	messages    chan ActionMessage

	// TODO RoomRepository to accept that correct user enters a room.
	rooms       map[uint64]*activeRoom
	clients     map[uint64]*activeClient
	activeConns map[*Client]activeConnection
}

func NewInitialRoom( /*RoomRepository*/ ) *InitialRoom {
	return &InitialRoom{
		connects:     make(chan *Client, 1),
		distconnects: make(chan *Client, 1),
		messages:     make(chan ActionMessage, 1),
		rooms:        make(map[uint64]*Room),
		clients:      make(map[uint64]*Client),
		activeConns:  make(map[*Client]activeConnection),
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

func (iroom *InitialRoom) connectClient(c *Client) {
	// add conncetion to active user when the user is already connected from anywhere.
	if activeC, ok := iroom.clients[c.userID]; ok {
		activeC.conns[c] = true
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
	c := NewClient(conn, u)
	c.onClosed = func(closedC *Client) {
		iroom.disconnects <- closedC
	}
	c.onAnyMessage = func(gotC *Client, msg interface{}) {
		m := msg.(ActionMessage)
		switch m.Action() {
		case ActionEnterRoom:
			m = enterRoomRequest{m, gotC}
		case ActionExitRoom:
			m = exitRoomRequest{m, gotC}
		}
		iroom.messages <- m
	}
	iroom.connects <- c
	c.Listen(ctx)
}

func (iroom *InitialRoom) disconnectClient(c *Client) {
	activeC, ok := iroom.clients[c.userID]
	if !ok {
		return
	}
	if _, ok = activeC.conns[c]; !ok {
		return
	}

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
		iroom.enterRoom(ctx, m.(EnterRoom))
	case ActionExitRoom:
		iroom.exitRoom(m.(ExitRoom))
	}
}

func (iroom *InitialRoom) enterRoom(ctx context.Context, req enterRoomRequest) {
	if err := iroom.validateClientHasRoom(req.Client, req.SenderID, req.RoomID); err != nil {
		sendError(req.Client, err)
		return
	}

	activeRoom, ok := iroom.rooms[req.RoomID]
	// if the room is inactive, activate it.
	if !ok {
		// TODO implement RoomRepository.getRoom()
		// roomEntity :=
		// activeRoom = newActiveRoom(NewRoom(roomEntity))
		// go activeRoom.room.Listen(ctx)
	}
	activeRoom.room.Join(req.Client)
	activeRoom.nMenber += 1
}

func (iroom *InitialRoom) exitRoom(req exitRoomRequest) {
	if err := iroom.validateClientHasRoom(req.Client, req.SenderID, req.RoomID); err != nil {
		sendError(req.Client, err)
		return
	}

	room, ok := iroom.rooms[req.RoomID]
	if !ok {
		log.Println("exit room for inactive room")
		return
	}
	room.room.Leave(req.Client)
	room.nMenber -= 1

	// expire the active room from which all members have been leaved.
	if room.nMenber == 0 {
		room.room.Done()
		delete(iroom.rooms, req.RoomID)
	}
}

func (iroom *InitialRoom) validateClientHasRoom(conn *Client, userID, roomID uint64) error {
	if !iroom.connectionExist(req.SenderID, req.Client) {
		return fmt.Errorf("request user id(%d) is not found", req.SenderID)
	}
	// the client is validated at above.
	if _, ok := iroom.clients[req.SenderID].rooms[req.RoomID]; !ok {
		return fmt.Errorf("user(%d) does not have request room id(%d) ", req.SenderID, req.RoomID)
	}
	return nil
}

// check whether active client with the websocket connection exists?
func (iroom *InitialRoom) connectionExist(userID uint64, conn *Client) bool {
	if activeC, ok := iroom.clients[req.SenderID]; ok {
		if _, ok := activeC.conns[req.Client]; ok {
			return true
		}
	}
	return false
}

func sendError(c *Client, err error) {
	log.Println(err)
	go func() { c.Send(NewErrorMessage(err)) }()
}
