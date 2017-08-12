package entity

import (
	"context"
)

type Room struct {
	ID         uint64
	Name       string
	IsTalkRoom bool
}

type RoomRepository interface {

	// get rooms which user has, from repository.
	GetUserRooms(ctx context.Context, userID uint64) ([]Room, error)

	// store new room to repository and return
	// stored room.
	Add(ctx context.Context, r Room) (Room, error)

	// remove room from repository.
	Remove(ctx context.Context, r Room) error

	Find(ctx context.Context, roomID uint64) (Room, error)

	// check whether the room specified by roomID has
	// member specified by userID.
	// return true if the room has the member.
	RoomHasMember(ctx context.Context, roomID, userID uint64) bool

	// Members(roomID uint64) ([]entity.User, error)
}
