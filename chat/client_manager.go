package chat

import (
	"context"
	"fmt"

	"github.com/shirasudon/go-chat/chat/action"
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
	conns   map[Conn]connectionInfo
	rooms   map[uint64]bool // has rooms managed by room id
	friends map[uint64]bool // has friends managed by user id
}

func newActiveClient(c Conn) *activeClient {
	return &activeClient{
		conns:   map[Conn]connectionInfo{c: connectionInfo{}},
		rooms:   make(map[uint64]bool),
		friends: make(map[uint64]bool),
	}
}

func (ac *activeClient) Send(m action.ActionMessage) {
	for c, _ := range ac.conns {
		c.Send(m)
	}
}

// ClientManager manages active clients.
type ClientManager struct {
	chatQuery *QueryService
	clients   map[uint64]*activeClient
}

func NewClientManager(service *QueryService) *ClientManager {
	return &ClientManager{
		chatQuery: service,
		clients:   make(map[uint64]*activeClient),
	}
}

func (cm *ClientManager) broadcastsFriends(ac *activeClient, m action.ActionMessage) {
	for friendID, _ := range ac.friends {
		if activeFriend, ok := cm.clients[friendID]; ok {
			activeFriend.Send(m)
		}
	}
}

func (cm *ClientManager) broadcastsUsers(userIDs []uint64, m action.ActionMessage) {
	for _, userID := range userIDs {
		if activeUser, ok := cm.clients[userID]; ok {
			activeUser.Send(m)
		}
	}
}

func (cm *ClientManager) connectClient(ctx context.Context, c Conn) error {
	// add conncetion to active user when the user is already connected from anywhere.
	if activeC, ok := cm.clients[c.UserID()]; ok {
		activeC.conns[c] = connectionInfo{}
		return nil
	}

	// create new active client because the connection is newly.
	// and broadcasts user connect event to all active friends
	activeC := newActiveClient(c)
	// set friends and rooms to new active user.
	relation, err := cm.chatQuery.FindUserRelation(ctx, c.UserID())
	if err != nil {
		return err
	}
	for _, f := range relation.Friends {
		activeC.friends[f.ID] = true
	}
	for _, r := range relation.Rooms {
		activeC.rooms[r.ID] = true
	}

	cm.clients[c.UserID()] = activeC

	// TODO replace by publishing domain event
	cm.broadcastsFriends(activeC, action.NewUserConnect(c.UserID()))

	return nil
}

func (cm *ClientManager) disconnectClient(c Conn) {
	activeC, ok := cm.clients[c.UserID()]
	if !ok {
		return
	}

	delete(activeC.conns, c)
	if len(activeC.conns) == 0 {
		delete(cm.clients, c.UserID())
		cm.broadcastsFriends(activeC, action.NewUserDisconnect(c.UserID()))
	}
}

func (cm *ClientManager) validateClientHasRoom(conn Conn, userID, roomID uint64) error {
	if !cm.connectionExist(userID, conn) {
		return fmt.Errorf("request user id(%d) is not found", userID)
	}
	// the client is validated at above.
	if _, ok := cm.clients[userID].rooms[roomID]; !ok {
		return fmt.Errorf("user(%d) does not have request room id(%d) ", userID, roomID)
	}
	return nil
}

// check whether active client with the websocket connection exists?
func (cm *ClientManager) connectionExist(userID uint64, conn Conn) bool {
	if activeC, ok := cm.clients[userID]; ok {
		if _, ok := activeC.conns[conn]; ok {
			return true
		}
	}
	return false
}