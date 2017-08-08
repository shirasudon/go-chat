package model

import (
	"context"
	"log"
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
	// TODO RoomRepository to accept that correct user enters a room.
	rooms map[uint64]*activeRoom
}

func NewRoomManager( /*RoomRepository*/ ) *RoomManager {
	return &RoomManager{
		rooms: make(map[uint64]*activeRoom),
	}
}

func (rm *RoomManager) Room(roomID uint64) *activeRoom {
	return rm.rooms[roomID]
}

func (rm *RoomManager) RoomExist(roomID uint64) bool {
	_, ok := rm.rooms[roomID]
	return ok
}

func (rm *RoomManager) EnterRoom(ctx context.Context, currentRoomID, nextRoomID uint64, c *Conn) {
	// TODO exit previous room before enter new room.
	if currentRoomID > 0 {
		rm.exitRoom(currentRoomID, c)
	}

	activeRoom, ok := rm.rooms[nextRoomID]
	// if the room is inactive, activate it.
	if !ok {
		// TODO implement RoomRepository.getRoom()
		// roomEntity :=
		// activeRoom = newActiveRoom(NewRoom(roomEntity))
		go activeRoom.room.Listen(ctx)
	}
	activeRoom.room.Join(c)
	activeRoom.nMenber += 1
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
