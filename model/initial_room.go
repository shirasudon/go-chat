package model

import (
	"context"
	"golang.org/x/net/websocket"
)

type clientState struct {
	havingRooms map[uint64]struct {
		room     *Room
		nClients int
	}
}

// InitialRoom is the room which have any clients created newly.
// any clients enters this room firstly, then dispatch their to requesting rooms.
// clients enter again this room after leaving the requesting rooms, then waiting for
// dispatch to next rooms they are requested.
//
// All of the rooms are children of this room. So They are managed by InitialRoom.
type InitialRoom struct {
	roomRequests chan UserJoinRoom // requesting to dispatch a room
	joins        chan *Client
	leaves       chan *Client
	messages     chan ActionMessage

	// TODO RoomRepository to accept that correct user enters a room.
	rooms        map[uint64]*Room
	clients      map[uint64]*Client
	clientStates map[*Client]clientState
}

func NewInitialRoom() {
	return &InitialRoom{
		roomRequests: make(chan UserJoinRoom, 1),
		joins:        make(chan *Client, 1),
		leaves:       make(chan *Client, 1),
		messages:     make(chan ActionMessage, 1),
		rooms:        make(map[uint64]*Room),
		clients:      make(map[uint64]*Client),
	}
}

func (iroom *InitialRoom) Listen(ctx context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	for {
		select {
		case req := <-iroom.roomRequests:
			iroom.dispatchRoom(ctx, req)
		case c := <-iroom.joins:
			iroom.joinClient(c)
		case c := <-iroom.leaves:
			iroom.leaveClient(c)
		case m := <-iroom.messages:
			iroom.handleMessage(m)
		case <-ctx.Done():
			return
		}
	}
}

func (iroom *InitialRoom) dispatchRoom(ctx context.Context, req UserJoinRoom) {
	c, ok := iroom.clients[req.SenderID]
	if !ok {
		// TODO error handling
		log.Println("no client exist")
		return
	}

	room, ok := iroom.rooms[req.RoomID]
	if !ok {
		// TODO check existance for room using userID and roomID
		c.Send(NewErrorMessage(fmt.Errorf("request room id(%d) is not found", req.RoomID)))
		room = NewRoom()
		go room.Listen(ctx)
	}
	room.Join(c)
}

func (iroom *InitialRoom) joinClient(c *Client) {
	iroom.clients[c.userID] = c
}

// Join new websocket client to room.
// it blocks until context is done.
func (iroom *InitialRoom) Join(ctx context.Context, conn *websocket.Conn, u entity.User) {
	c := NewClient(conn, u)
	c.onClosed = func(closedC *Client) {
		iroom.leaves <- closedC
	}
	c.onAnyMessage = func(gotC *Client, msg interface{}) {
		iroom.messages <- msg
	}
	iroom.joins <- c
	c.Listen(ctx)
}

func (iroom *InitialRoom) leaveClient(c *Client) {
	delete(iroom.clients, c.userID)
}

func (iroom *InitialRoom) handleMessage(m ActionMessage) {
	switch m.Action() {
	case ActionCreateRoom:
	case ActionDeleteRoom:
	}
}
