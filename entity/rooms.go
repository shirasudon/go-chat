package entity

import (
	"context"
	"errors"
	"fmt"
)

type Room struct {
	ID         uint64
	Name       string
	IsTalkRoom bool
}

type RoomRepository interface {
	// get one room.
	Find(ctx context.Context, roomID uint64) (Room, error)

	// get all the rooms which user has, from repository.
	FindAllByUserID(ctx context.Context, userID uint64) ([]Room, error)

	// store new room to repository and return
	// stored room id.
	Store(ctx context.Context, r Room) (uint64, error)

	// remove room from repository.
	Remove(ctx context.Context, roomID uint64) error
}

type RoomRelation struct {
	Room
	Members  map[uint64]User
	Messages []Message // TODO cache to avoid a large allocation.
}

func (rr *RoomRelation) AddMessage(userID uint64, content string) error {
	// non room member's message is invalid.
	if _, ok := rr.Members[userID]; !ok {
		return fmt.Errorf("user(id=%d) is not a member in the room(id=%d)", userID, rr.ID)
	}
	msg := Message{}
	msg.RoomID = rr.ID
	msg.UserID = userID
	msg.Content = content
	rr.Messages = append(rr.Messages, msg)
	return nil
}

func (rr *RoomRelation) AddMember(u User) error {
	// the user not in the datastore is invalid.
	if u.ID < 1 {
		return errors.New("invalid user")
	}
	// the user already exist in the room is invalid
	if _, ok := rr.Members[u.ID]; ok {
		return fmt.Errorf("user(id=%d) is already member in the room(id=%d)", u.ID, rr.ID)
	}
	rr.Members[u.ID] = u
	return nil
}

type RoomRelationRepository interface {
	Find(ctx context.Context, roomID uint64) (RoomRelation, error)

	// check whether the room specified by roomID has
	// member specified by userID.
	// return true if the room has the member.
	ExistRoomMember(ctx context.Context, roomID, userID uint64) bool
}
