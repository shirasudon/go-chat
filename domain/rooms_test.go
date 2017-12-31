package domain

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/shirasudon/go-chat/domain/event"
)

type RoomRepositoryStub struct{}

func (r *RoomRepositoryStub) BeginTx(context.Context, *sql.TxOptions) (Tx, error) {
	panic("not implemented")
}

func (r *RoomRepositoryStub) Find(ctx context.Context, roomID uint64) (Room, error) {
	panic("not implemented")
}

func (r *RoomRepositoryStub) FindAllByUserID(ctx context.Context, userID uint64) ([]Room, error) {
	panic("not implemented")
}

func (rr *RoomRepositoryStub) Store(ctx context.Context, r Room) (uint64, error) {
	return r.ID + 1, nil
}

func (rr *RoomRepositoryStub) Remove(ctx context.Context, r Room) error {
	return nil
}

var roomRepo = &RoomRepositoryStub{}

func TestRoomCreated(t *testing.T) {
	ctx := context.Background()
	owner := &User{ID: 3}
	r, err := NewRoom(ctx, roomRepo, "test", owner, NewUserIDSet(1))
	if err != nil {
		t.Fatal(err)
	}

	if r.ID == 0 {
		t.Fatalf("room is created but has invalid ID(%d)", r.ID)
	}
	if !r.MemberIDSet.Has(owner.ID) {
		t.Fatalf("created room does not have owner id as room member, expect id: %d", owner.ID)
	}
	if r.CreatedAt == (time.Time{}) {
		t.Errorf("created room does not have created time, got: %v", r.CreatedAt)
	}
	if read, ok := r.MemberReadTimes[owner.ID]; !ok || read == (time.Time{}) {
		t.Errorf("created room missed read time for owner of the Room")
	}

	// check whether room has one event,
	events := r.Events()
	if got := len(events); got != 1 {
		t.Errorf("room has no event after RoomCreated")
	}
	ev, ok := events[0].(event.RoomCreated)
	if !ok {
		t.Errorf("invalid event state for the room")
	}

	// check whether room created event is valid.
	if got := ev.Name; got != "test" {
		t.Errorf("RoomCreated has different room name, expect: %s, got: %s", "test", got)
	}
	if got := ev.RoomID; got != r.ID {
		t.Errorf("RoomCreated has different room id, expect: %v, got: %v", r.ID, got)
	}
	if got, expect := len(ev.MemberIDs), len(r.MemberIDs()); got != expect {
		t.Errorf("RoomCreated has dieffrent room members size, expect: %d, got: %d", expect, got)
	}
	if got := ev.Timestamp(); got == (time.Time{}) {
		t.Error("RoomCreated has no timestamp")
	}
}

func TestRoomDeletedSuccess(t *testing.T) {
	ctx := context.Background()
	owner := &User{ID: 3}
	r, err := NewRoom(ctx, roomRepo, "test", owner, NewUserIDSet(1))
	if err != nil {
		t.Fatal(err)
	}

	var deletedID = r.ID
	err = r.Delete(ctx, roomRepo, owner)
	if err != nil {
		t.Fatalf("can not be deleted the room")
	}

	if !r.NotExist() {
		t.Fatalf("room is deleted but has invalid ID(%d)", r.ID)
	}

	// check whether room has deleted event,
	// currently two events: created and deleted.
	events := r.Events()
	if got := len(events); got != 2 {
		t.Errorf("room has invalid event after RoomCreated and RoomDeleted")
	}
	ev, ok := events[1].(event.RoomDeleted)
	if !ok {
		t.Errorf("invalid event state for the room")
	}

	// check whether room deleted event is valid.
	if got := ev.RoomID; got != deletedID {
		t.Errorf("RoomDeleted has different room id, expect: %d, got: %d", deletedID, got)
	}
	if got := ev.Name; got != r.Name {
		t.Errorf("RoomDeleted has different room name, expect: %s, got: %s", r.Name, got)
	}
	if got, expect := len(ev.MemberIDs), len(r.MemberIDs()); got != expect {
		t.Errorf("RoomDeleted has dieffrent room members size, expect: %d, got: %d", expect, got)
	}
	if got := ev.Timestamp(); got == (time.Time{}) {
		t.Error("RoomDeleted has no timestamp")
	}
}

func TestRoomDeletedFail(t *testing.T) {
	ctx := context.Background()
	err := (&Room{}).Delete(ctx, roomRepo, &User{})
	if err == nil {
		t.Fatalf("the room not in the datastore is deleted")
	}
}

func TestRoomAddMember(t *testing.T) {
	ctx := context.Background()
	owner := &User{ID: 3}
	r, _ := NewRoom(ctx, roomRepo, "test", owner, NewUserIDSet())
	r.ID = 1 // it may not be allowed at application side.
	u := User{ID: 1}
	ev, err := r.AddMember(u)
	if err != nil {
		t.Fatal(err)
	}
	if got := ev.RoomID; got != r.ID {
		t.Errorf("RoomAddedMember has different room id, expect: %d, got: %d", r.ID, got)
	}
	if got := ev.AddedUserID; got != u.ID {
		t.Errorf("RoomAddedMember has different user id, expect: %d, got: %d", u.ID, got)
	}
	if got := ev.Timestamp(); got == (time.Time{}) {
		t.Error("RoomAddedMember has no timestamp")
	}

	if !r.HasMember(u) {
		t.Errorf("AddMember does not add any member to the room")
	}
	if _, ok := r.MemberReadTimes[u.ID]; !ok {
		t.Errorf("AddMember missed read time of the added member")
	}

	// room has two events: Created, AddedMember.
	if got := len(r.Events()); got != 2 {
		t.Errorf("room has no event")
	}
	if _, ok := r.Events()[1].(event.RoomAddedMember); !ok {
		t.Errorf("invalid event is added")
	}
}

func TestRoomRemoveMember(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	owner := &User{ID: 3}
	r, _ := NewRoom(ctx, roomRepo, "test", owner, NewUserIDSet())
	r.ID = 1 // it may not be allowed at application side.
	u := User{ID: 1}
	_, err := r.AddMember(u)
	if err != nil {
		t.Fatal(err)
	}

	// do test function
	ev, err := r.RemoveMember(u)
	if err != nil {
		t.Fatal(err)
	}

	if got := ev.RoomID; got != r.ID {
		t.Errorf("RoomRemovedMember has different room id, expect: %d, got: %d", r.ID, got)
	}
	if got := ev.RemovedUserID; got != u.ID {
		t.Errorf("RoomRemovedMember has different removed user id, expect: %d, got: %d", u.ID, got)
	}
	if got := ev.Timestamp(); got == (time.Time{}) {
		t.Error("RoomRemovedMember has no timestamp")
	}

	if r.HasMember(u) {
		t.Errorf("RemoveMember does not remove a member from the room")
	}
	if _, ok := r.MemberReadTimes[u.ID]; ok {
		t.Errorf("RemoveMember missed removing read time of the removed member")
	}

	// room has three events: Created, AddedMember, RemovedMember
	if got := len(r.Events()); got != 3 {
		t.Errorf("room has different events")
	}
	gotEv, ok := r.Events()[2].(event.RoomRemovedMember)
	if !ok {
		t.Fatalf("invalid event is added, expect: %T, got: %T", event.RoomRemovedMember{}, gotEv)
	}
	if ev != gotEv {
		t.Errorf("different event fields on RoomRemovedMember")
	}
}

func TestRoomRemoveMemberFail(t *testing.T) {
	t.Parallel()

	{ // case1: removed member is not found.
		ctx := context.Background()
		owner := &User{ID: 3}
		r, _ := NewRoom(ctx, roomRepo, "test", owner, NewUserIDSet())
		r.ID = 1 // it may not be allowed at application side.
		u := User{ID: 1}
		_, err := r.RemoveMember(u)
		if err == nil {
			t.Error("removed not found user, but no error")
		}
		// currently events: created only
		if got := len(r.Events()); got != 1 {
			t.Error("RemoveMember failed, but event is added to the room")
		}
	}
}

func TestRoomReadMessagesByUser(t *testing.T) {
	ctx := context.Background()
	owner := &User{ID: 3}
	r, _ := NewRoom(ctx, roomRepo, "test", owner, NewUserIDSet())
	r.ID = 1 // it may not be allowed at application side.

	// case1: success.
	{
		now := time.Now()
		ev, err := r.ReadMessagesBy(owner, now)
		if err != nil {
			t.Fatal(err)
		}
		if got := ev.RoomID; got != r.ID {
			t.Errorf("MessageReadByUser has different room id, expect: %d, got: %d", r.ID, got)
		}
		if got := ev.UserID; got != owner.ID {
			t.Errorf("MessageReadByUser has different user id, expect: %d, got: %d", owner.ID, got)
		}
		if got := ev.Timestamp(); got == (time.Time{}) {
			t.Error("MessageReadByUser has no timestamp")
		}

		if got := r.MemberReadTimes[owner.ID]; !got.Equal(now) {
			t.Errorf("ReadMessagesByUser does not set new read time, expect: %v, got: %v", now, got)
		}

		// room has two events: Created, MessageReadByUser
		if got := len(r.Events()); got != 2 {
			t.Errorf("room has no event")
		}
		if _, ok := r.Events()[1].(event.RoomMessagesReadByUser); !ok {
			t.Errorf("invalid event is added")
		}
	}

	// case2: past read
	{
		past := time.Now().Add(-24 * time.Hour)
		_, err := r.ReadMessagesBy(owner, past)
		if err == nil {
			t.Error("passing past time as read time, but no error")
		}
	}

	// case3: read by not a room member
	{
		const NotExistUserID = uint64(999)
		_, err := r.ReadMessagesBy(&User{ID: NotExistUserID}, time.Now())
		if err == nil {
			t.Error("read by not a room member, but no error")
		}
	}
}

func TestGetSetReadTime(t *testing.T) {
	r := Room{}
	_, ok := r.getMemberReadTime(1)
	if ok {
		t.Error("no entry but get returned ok")
	}

	now := time.Now()
	r.setMemberReadTime(1, now)

	readTime, ok := r.getMemberReadTime(1)
	if !ok {
		t.Error("set entry but get returned not-ok")
	}
	if !readTime.Equal(now) {
		t.Error("different time")
	}
}
