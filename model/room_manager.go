package model

import (
	"context"
	"log"

	"github.com/shirasudon/go-chat/entity"
)

// activeRoom is a wrapper for Room which has
// a number of active chat members.
type activeRoom struct {
	room    *Room
	nMenber int
}

func newActiveRoom(r *Room) *activeRoom {
	return &activeRoom{
		room:    r,
		nMenber: 0,
	}
}

type RoomManager struct {
	roomRepo entity.RoomRepository
	msgRepo  entity.MessageRepository
	rooms    map[uint64]*activeRoom
}

func NewRoomManager(repos entity.Repositories) *RoomManager {
	return &RoomManager{
		roomRepo: repos.Rooms(),
		msgRepo:  repos.Messages(),
		rooms:    make(map[uint64]*activeRoom),
	}
}

func (rm *RoomManager) Room(roomID uint64) *activeRoom {
	return rm.rooms[roomID]
}

func (rm *RoomManager) RoomExist(roomID uint64) bool {
	_, ok := rm.rooms[roomID]
	return ok
}

func (rm *RoomManager) EnterRoom(ctx context.Context, currentRoomID, nextRoomID uint64, c *Conn) error {
	// exit previous room before enter new room.
	if currentRoomID > 0 {
		rm.exitRoom(currentRoomID, c)
	}

	activeRoom, ok := rm.rooms[nextRoomID]
	// if the room is inactive, activate it.
	if !ok {
		roomEntity, err := rm.roomRepo.Find(ctx, nextRoomID)
		if err != nil {
			return err
		}
		room := NewRoom(roomEntity, rm.msgRepo)
		activeRoom = newActiveRoom(room)
		go activeRoom.room.Listen(ctx)
	}
	activeRoom.room.Join(c)
	activeRoom.nMenber += 1
	return nil
}

func (rm *RoomManager) exitRoom(roomID uint64, c *Conn) {
	room, ok := rm.rooms[roomID]
	if !ok {
		log.Println("exit room for inactive room")
		return
	}
	room.room.Leave(c)
	room.nMenber -= 1

	// expire the active room from which all members have been leaved.
	if room.nMenber == 0 {
		room.room.Done()
		delete(rm.rooms, roomID)
	}
}

func (rm *RoomManager) DisconnectClient(currentRoomID uint64, c *Conn) {
	rm.exitRoom(currentRoomID, c)
}

func (rm *RoomManager) Send(m ToRoomMessage) {
	activeR, ok := rm.rooms[m.ToRoom()]
	if !ok {
		return
	}
	activeR.room.Send(m)
}
