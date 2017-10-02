package entity

import (
	"context"
	"fmt"
)

type Room struct {
	ID         uint64
	Name       string
	IsTalkRoom bool

	MemberIDs map[uint64]bool

	MessageIDs []uint64 // TODO cache to avoid a large allocation.
}

func (r *Room) getMemberIDs() map[uint64]bool {
	if r.MemberIDs == nil {
		r.MemberIDs = make(map[uint64]bool, 4)
	}
	return r.MemberIDs
}

func (r *Room) AddMessage(ctx context.Context, msgRepo MessageRepository, userID uint64, content string) error {
	// non room member's message is invalid.
	if _, ok := r.getMemberIDs()[userID]; !ok {
		return fmt.Errorf("user(id=%d) is not a member in the room(id=%d)", userID, r.ID)
	}
	msg := Message{}
	msg.RoomID = r.ID
	msg.UserID = userID
	msg.Content = content

	msgID, err := msgRepo.Add(ctx, msg)
	if err != nil {
		return err
	}
	r.MessageIDs = append(r.MessageIDs, msgID)
	return nil
}

func (r *Room) AddMember(ctx context.Context, userRepo UserRepository, userID uint64) error {
	// the user not shown in the datastore is invalid.
	if _, err := userRepo.Find(ctx, userID); err != nil {
		return fmt.Errorf("user(id=%d) does not exist in the datastore", userID)
	}
	// the user already exist in the room is invalid
	if _, ok := r.getMemberIDs()[userID]; ok {
		return fmt.Errorf("user(id=%d) is already member in the room(id=%d)", userID, r.ID)
	}
	r.MemberIDs[userID] = true
	return nil
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
