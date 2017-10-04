package entity

import (
	"context"
	"fmt"
)

type Room struct {
	ID         uint64
	Name       string
	IsTalkRoom bool

	// user set, order has no meaning
	memberIDs map[uint64]bool

	// message list, in order older messages;
	// first index has the oldest message and last has the lastest.
	messageIDs []uint64
}

// It returns a deep copy of the room member's IDs.
func (r *Room) MemberIDs() []uint64 {
	memberMap := r.getMemberIDs()
	memberIDs := make([]uint64, 0, len(memberMap))
	for id, _ := range memberMap {
		memberIDs = append(memberIDs, id)
	}
	return memberIDs
}

func (r *Room) getMemberIDs() map[uint64]bool {
	if r.memberIDs == nil {
		r.memberIDs = make(map[uint64]bool, 4)
	}
	return r.memberIDs
}

// It adds the member specified by userID to the room.
// It returns error when the user already exist in the room.
func (r *Room) AddMember(userID uint64) error {
	if r.HasMember(userID) {
		return fmt.Errorf("user(id=%d) is already member of the room(id=%d)", userID, r.ID)
	}
	r.getMemberIDs()[userID] = true
	return nil
}

// It returns true when the room member specified by userID exist
// in the room, otherwise returns false.
func (r *Room) HasMember(userID uint64) bool {
	_, exist := r.getMemberIDs()[userID]
	return exist
}

// It adds the chat message to the room.
// It returns error when the userID which send the message is not found
// in the room, or something happened as store new message to the repository.
//
// TODO make msgRepo is independent?
func (r *Room) AddMessage(ctx context.Context, msgRepo MessageRepository, userID uint64, content string) error {
	// non room member's message is invalid.
	if !r.HasMember(userID) {
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
	r.messageIDs = append(r.messageIDs, msgID)
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
