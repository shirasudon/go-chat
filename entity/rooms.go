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

func NewRoom(id uint64, name string, memberIDs map[uint64]bool) Room {
	return Room{
		ID:         id,
		Name:       name,
		IsTalkRoom: false,
		memberIDs:  memberIDs,
		messageIDs: make([]uint64, 0, 4),
	}
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

// It adds the member to the room.
// It returns action event and error when the user already exist in the room.
func (r *Room) AddMember(user User) (RoomAddedMember, error) {
	if r.HasMember(user) {
		return RoomAddedMember{}, fmt.Errorf("user(id=%d) is already member of the room(id=%d)", user.ID, r.ID)
	}

	r.getMemberIDs()[user.ID] = true

	return RoomAddedMember{
		RoomID:      r.ID,
		AddedUserID: user.ID,
	}, nil
}

// It returns true when the room member exists
// in the room, otherwise returns false.
func (r *Room) HasMember(member User) bool {
	_, exist := r.getMemberIDs()[member.ID]
	return exist
}

// It adds the chat message to the room.
// It returns action event and error when the user which send the message is not found
// in the room.
func (r *Room) PostMessage(user User, content string) (RoomPostedMessage, error) {
	// non room member's message is invalid.
	if !r.HasMember(user) {
		return RoomPostedMessage{}, fmt.Errorf("user(id=%d) is not a member in the room(id=%d)", user.ID, r.ID)
	}
	return RoomPostedMessage{
		PostUserID:   user.ID,
		PostedRoomID: r.ID,
		Content:      content,
	}, nil
}

type RoomRepository interface {
	TxBeginner

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
