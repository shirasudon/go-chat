package entity

import (
	"context"
	"fmt"
)

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

// Room entity. Its fields are exported
// due to construct from the datastore.
// In application side, creating/modifying/deleting the room
// should be done by the methods which emits the domain event.
type Room struct {
	EventHolder

	// ID = 0 means new entity which is not in the repository
	ID         uint64
	Name       string
	IsTalkRoom bool

	// user set, order has no meaning
	MemberIDSet map[uint64]bool
}

// create new Room entity. the retruned room holds RoomCreated event
// which also returns the second result.
func NewRoom(name string, memberIDs map[uint64]bool) (Room, RoomCreated) {
	if memberIDs == nil {
		memberIDs = make(map[uint64]bool, 2)
	}

	r := Room{
		EventHolder: NewEventHolder(),
		ID:          0, // 0 means new entity
		Name:        name,
		IsTalkRoom:  false,
		MemberIDSet: memberIDs,
	}
	ev := RoomCreated{
		Name:       name,
		IsTalkRoom: false,
		MemberIDs:  r.MemberIDs(),
	}
	r.AddEvent(ev)
	return r, ev // TODO event should be returned?
}

// It returns a deep copy of the room member's IDs as list.
func (r *Room) MemberIDs() []uint64 {
	memberMap := r.getMemberIDSet()
	memberIDs := make([]uint64, 0, len(memberMap))
	for id, _ := range memberMap {
		memberIDs = append(memberIDs, id)
	}
	return memberIDs
}

func (r *Room) getMemberIDSet() map[uint64]bool {
	if r.MemberIDSet == nil {
		r.MemberIDSet = make(map[uint64]bool, 4)
	}
	return r.MemberIDSet
}

// It adds the member to the room.
// It returns the event adding to the room, and error
// when the user already exist in the room.
func (r *Room) AddMember(user User) (RoomAddedMember, error) {
	if r.ID == 0 {
		return RoomAddedMember{}, fmt.Errorf("newly room can not be added new member")
	}
	if r.HasMember(user) {
		return RoomAddedMember{}, fmt.Errorf("user(id=%d) is already member of the room(id=%d)", user.ID, r.ID)
	}

	r.getMemberIDSet()[user.ID] = true

	ev := RoomAddedMember{
		RoomID:      r.ID,
		AddedUserID: user.ID,
	}
	r.AddEvent(ev)
	return ev, nil
}

// It returns true when the room member exists
// in the room, otherwise returns false.
func (r *Room) HasMember(member User) bool {
	_, exist := r.getMemberIDSet()[member.ID]
	return exist
}

// It adds the chat message to the room.
// It returns action event and error when the user which send the message is not found
// in the room.
func (r *Room) PostMessage(user User, content string) (RoomPostedMessage, error) {
	if r.ID == 0 {
		return RoomPostedMessage{}, fmt.Errorf("newly room can not be posted new message")
	}
	// non room member's message is invalid.
	if !r.HasMember(user) {
		return RoomPostedMessage{}, fmt.Errorf("user(id=%d) is not a member in the room(id=%d)", user.ID, r.ID)
	}

	ev := RoomPostedMessage{
		PostUserID:   user.ID,
		PostedRoomID: r.ID,
		Content:      content,
	}
	r.AddEvent(ev)
	return ev, nil
}
