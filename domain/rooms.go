package domain

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/shirasudon/go-chat/domain/event"
)

//go:generate mockgen -destination=../internal/mocks/mock_rooms.go -package=mocks github.com/shirasudon/go-chat/domain RoomRepository

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

	CreatedAt time.Time

	OwnerID     uint64
	MemberIDSet UserIDSet

	// key: userID, value: ReadTime
	MemberReadTimes TimeSet
}

// TimeSet is a set for the time.Time.
// empty value is valid for use.
type TimeSet struct {
	set map[uint64]time.Time
}

func NewTimeSet(ids ...uint64) TimeSet {
	set := make(map[uint64]time.Time, len(ids))
	for _, id := range ids {
		set[id] = time.Time{}
	}
	return TimeSet{
		set: set,
	}
}

func (set *TimeSet) getMap() map[uint64]time.Time {
	if set.set == nil {
		set.set = make(map[uint64]time.Time, 8) // TODO arbitrary value 8
	}
	return set.set
}

// Get returns time.Time corresponding to the id and true.
// If id not found It's second returned value is false.
func (set *TimeSet) Get(id uint64) (time.Time, bool) {
	t, ok := set.getMap()[id]
	return t, ok
}

// Set sets time.Time with corresponding id to the internal map.
func (set *TimeSet) Set(id uint64, t time.Time) {
	set.getMap()[id] = t
}

// Delete deletes id from the internal map.
func (set *TimeSet) Delete(id uint64) {
	delete(set.getMap(), id)
}

// create new Room entity into the repository. It retruns room holding RoomCreated event
// and error if any.
func NewRoom(ctx context.Context, roomRepo RoomRepository, name string, user *User, memberIDs UserIDSet) (*Room, error) {
	if user.NotExist() {
		return nil, fmt.Errorf("the user not in the datastore, can not create room")
	}

	// room owner should be contain in the MemberIDSet.
	if !memberIDs.Has(user.ID) {
		memberIDs.Add(user.ID)
	}

	now := time.Now()
	timeSet := NewTimeSet()
	for id, _ := range memberIDs.idMap {
		timeSet.Set(id, now)
	}

	r := &Room{
		EventHolder:     NewEventHolder(),
		ID:              0, // 0 means new entity
		Name:            name,
		IsTalkRoom:      false,
		CreatedAt:       now,
		OwnerID:         user.ID,
		MemberIDSet:     memberIDs,
		MemberReadTimes: timeSet,
	}
	id, err := roomRepo.Store(ctx, *r)
	if err != nil {
		return nil, err
	}

	r.ID = id

	ev := event.RoomCreated{
		CreatedBy:  user.ID,
		RoomID:     id,
		Name:       name,
		IsTalkRoom: false,
		MemberIDs:  memberIDs.List(),
	}
	ev.Occurs()
	r.AddEvent(ev)

	return r, nil
}

// It deletes the room from repository.
// After successing that, the room holds RoomDeleted event.
func (r *Room) Delete(ctx context.Context, repo RoomRepository, user *User) error {
	if r.NotExist() {
		return fmt.Errorf("the room not in the datastore, can not be deleted")
	}
	if user.NotExist() {
		return fmt.Errorf("the user not in the datastore, can not delete the room")
	}
	if r.OwnerID != user.ID {
		return fmt.Errorf("the user is not the owner of the room, can not delete the room")
	}

	err := repo.Remove(ctx, *r)
	if err != nil {
		return err
	}

	removedID := r.ID
	r.ID = 0 // means not in the repository.

	ev := event.RoomDeleted{
		RoomID:     removedID,
		DeletedBy:  user.ID,
		Name:       r.Name,
		IsTalkRoom: r.IsTalkRoom,
		MemberIDs:  r.MemberIDs(),
	}
	ev.Occurs()
	r.AddEvent(ev)

	return nil
}

// It returns whether the room is newly.
func (r *Room) NotExist() bool {
	return r == nil || r.ID == 0
}

// It returns a deep copy of the room member's IDs as list.
func (r *Room) MemberIDs() []uint64 {
	return r.MemberIDSet.List()
}

// It adds the member to the room.
// It returns the event adding to the room, and error
// when the user already exist in the room.
func (r *Room) AddMember(user User) (event.RoomAddedMember, error) {
	if r.NotExist() {
		return event.RoomAddedMember{}, fmt.Errorf("newly room can not be added new member")
	}
	if user.NotExist() {
		return event.RoomAddedMember{}, fmt.Errorf("the user not in the datastore, can not be a room member")
	}
	if r.HasMember(user) {
		return event.RoomAddedMember{}, fmt.Errorf("user(id=%d) is already member of the room(id=%d)", user.ID, r.ID)
	}

	r.MemberIDSet.Add(user.ID)
	r.MemberReadTimes.Set(user.ID, r.CreatedAt)

	ev := event.RoomAddedMember{
		RoomID:      r.ID,
		AddedUserID: user.ID,
	}
	ev.Occurs()
	r.AddEvent(ev)
	return ev, nil
}

// It returns true when the room member exists
// in the room, otherwise returns false.
func (r *Room) HasMember(member User) bool {
	return r.MemberIDSet.Has(member.ID)
}

// RoomRemovedMember removes the member from the room.
// It returns the event the room member is removed or
// returns error when the user already does not exist in the room.
func (r *Room) RemoveMember(user User) (event.RoomRemovedMember, error) {
	if r.NotExist() {
		return event.RoomRemovedMember{}, fmt.Errorf("newly room can not remove a member")
	}
	if user.NotExist() {
		return event.RoomRemovedMember{}, fmt.Errorf("the user not in the datastore, can not be removed from the room")
	}
	if !r.HasMember(user) {
		return event.RoomRemovedMember{}, fmt.Errorf("user(id=%d) is not a member of the room(id=%d)", user.ID, r.ID)
	}
	if r.OwnerID == user.ID {
		return event.RoomRemovedMember{}, fmt.Errorf("the room owner(id=%d) can not removed from the room(id=%d)", r.OwnerID, r.ID)
	}

	r.MemberIDSet.Remove(user.ID)
	r.MemberReadTimes.Delete(user.ID)

	ev := event.RoomRemovedMember{
		RoomID:        r.ID,
		RemovedUserID: user.ID,
	}
	ev.Occurs()
	r.AddEvent(ev)
	return ev, nil
}

// ReadMessagesBy marks that the room messages before time readAt
// are read by the specified user.
//
// It returns MessageReadByUser event and error if any.
func (r *Room) ReadMessagesBy(u *User, readAt time.Time) (event.RoomMessagesReadByUser, error) {
	if r.NotExist() {
		return event.RoomMessagesReadByUser{}, errors.New("newly room can not be read messages by user")
	}
	if u.NotExist() {
		return event.RoomMessagesReadByUser{}, errors.New("the user not in the datastore, can not read any message")
	}
	if !r.HasMember(*u) {
		return event.RoomMessagesReadByUser{}, fmt.Errorf("user (id=%d) is not a member of the room (id=%d)", u.ID, r.ID)
	}

	// TODO raise error if the messages between prevRead and readAt not exist.
	// how we check the existance of the messages then?
	prevRead, ok := r.MemberReadTimes.Get(u.ID)
	if !ok {
		// room has a member with u.ID, but ReadTime not set.
		// reset default here.
		prevRead = r.CreatedAt
	}
	if prevRead.Equal(readAt) || prevRead.After(readAt) {
		return event.RoomMessagesReadByUser{}, fmt.Errorf("message read time (%v) must be after previous read time (%v)", readAt.Format(time.Stamp), prevRead.Format(time.Stamp))
	}
	r.MemberReadTimes.Set(u.ID, readAt)

	ev := event.RoomMessagesReadByUser{
		UserID: u.ID,
		RoomID: r.ID,
		ReadAt: readAt,
	}
	ev.Occurs()
	r.AddEvent(ev)
	return ev, nil
}
