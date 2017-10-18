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

// create new Room entity into the repository. It retruns room holding RoomCreated event
// and error if any.
func NewRoom(ctx context.Context, roomRepo RoomRepository, name string, memberIDs UserIDSet) (Room, error) {
	r := Room{
		EventHolder: NewEventHolder(),
		ID:          0, // 0 means new entity
		Name:        name,
		IsTalkRoom:  false,
		MemberIDSet: memberIDs,
	}
	id, err := roomRepo.Store(ctx, r)
	if err != nil {
		return r, err
	}

	r.ID = id

	ev := RoomCreated{
		Name:       name,
		IsTalkRoom: false,
		MemberIDs:  memberIDs.List(),
	}
	r.AddEvent(ev)

	return r, nil
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

// -----------------------
// Room events
// -----------------------

// Event for Room is created.
type RoomCreated struct {
	Name       string
	IsTalkRoom bool
	MemberIDs  []uint64
}

func (RoomCreated) EventType() EventType { return EventRoomCreated }

// Event for Room is deleted.
type RoomDeleted struct {
	RoomID     uint64
	Name       string
	IsTalkRoom bool
	MemberIDs  []uint64
}

func (RoomDeleted) EventType() EventType { return EventRoomDeleted }

// Event for Room added new member.
type RoomAddedMember struct {
	RoomID      uint64
	AddedUserID uint64
}

func (RoomAddedMember) EventType() EventType { return EventRoomAddedMember }
