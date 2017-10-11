package domain

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
	Remove(ctx context.Context, r Room) error
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

	MemberIDSet UserIDSet
}

// create new Room entity. the retruned room holds RoomCreated event
// which also returns the second result.
func NewRoom(name string, memberIDs UserIDSet) (Room, RoomCreated) {
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
		MemberIDs:  memberIDs.List(),
	}
	r.AddEvent(ev)
	return r, ev // TODO event should be returned?
}

// It returns whether the room is newly.
func (r *Room) IsNew() bool {
	return r.ID == 0
}

// It returns a deep copy of the room member's IDs as list.
func (r *Room) MemberIDs() []uint64 {
	return r.MemberIDSet.List()
}

// It adds the member to the room.
// It returns the event adding to the room, and error
// when the user already exist in the room.
func (r *Room) AddMember(user User) (RoomAddedMember, error) {
	if r.ID == 0 {
		return RoomAddedMember{}, fmt.Errorf("newly room can not be added new member")
	}
	if user.ID == 0 {
		return RoomAddedMember{}, fmt.Errorf("the user not in the datastore, can not be a room member")
	}
	if r.HasMember(user) {
		return RoomAddedMember{}, fmt.Errorf("user(id=%d) is already member of the room(id=%d)", user.ID, r.ID)
	}

	r.MemberIDSet.Add(user.ID)

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
	return r.MemberIDSet.Has(member.ID)
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
