package model

import (
	"context"

	"github.com/shirasudon/go-chat/entity"
)

// activeRoom is a wrapper for entity.Room which has
// a number of active members.
type activeRoom struct {
	entity.Room

	// room members in the Repository, including both active and inactive client.
	members map[uint64]bool

	// the number of active connections
	nActiveMembers int
}

func newActiveRoom(r entity.Room) *activeRoom {
	memberIDs := r.MemberIDs()
	members := make(map[uint64]bool, len(memberIDs))
	for _, id := range memberIDs {
		members[id] = true
	}
	return &activeRoom{
		Room:           r,
		members:        members,
		nActiveMembers: 0,
	}
}

type RoomManager struct {
	chatService *ChatService
	rooms       map[uint64]*activeRoom
}

func NewRoomManager(service *ChatService) *RoomManager {
	return &RoomManager{
		chatService: service,
		rooms:       make(map[uint64]*activeRoom),
	}
}

func (rm *RoomManager) roomMemberIDs(roomID uint64) []uint64 {
	activeR, ok := rm.rooms[roomID]
	if !ok {
		return []uint64{}
	}

	ids := make([]uint64, len(activeR.members))
	for id, _ := range activeR.members {
		ids = append(ids, id)
	}
	return ids
}

func (rm *RoomManager) connectClient(ctx context.Context, userID uint64) error {
	ur, err := rm.chatService.FindUserRelation(ctx, userID)
	if err != nil {
		return err
	}

	// increase active member count for active rooms.
	// or activate room if inactive.
	for _, r := range ur.Rooms {
		activeR, ok := rm.rooms[r.ID]
		if !ok {
			activeR = newActiveRoom(r)
			rm.rooms[r.ID] = activeR
		}
		activeR.nActiveMembers += 1
	}
	return nil
}

func (rm *RoomManager) disconnectClient(ctx context.Context, userID uint64) error {
	ur, err := rm.chatService.FindUserRelation(ctx, userID)
	if err != nil {
		return err
	}

	// decrease active member count for active rooms.
	// and expires if no member exist
	for _, r := range ur.Rooms {
		activeR, ok := rm.rooms[r.ID]
		if !ok {
			continue
		}
		activeR.nActiveMembers -= 1
		if activeR.nActiveMembers == 0 {
			delete(rm.rooms, r.ID)
		}
	}
	return nil
}
