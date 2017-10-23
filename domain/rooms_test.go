package domain

import (
	"context"
	"database/sql"
	"testing"
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

	// check whether room has one event,
	events := r.Events()
	if got := len(events); got != 1 {
		t.Errorf("room has no event after RoomCreated")
	}
	ev, ok := events[0].(RoomCreated)
	if !ok {
		t.Errorf("invalid event state for the room")
	}

	// check whether room created event is valid.
	if got := ev.Name; got != "test" {
		t.Errorf("RoomCreated has different room name, expect: %s, got: %s", "test", got)
	}
	if got, expect := len(ev.MemberIDs), len(r.MemberIDs()); got != expect {
		t.Errorf("RoomCreated has dieffrent room members size, expect: %d, got: %d", expect, got)
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
	ev, ok := events[1].(RoomDeleted)
	if !ok {
		t.Errorf("invalid event state for the room")
	}

	// check whether room deleted event is valid.
	if got := ev.RoomID; got != deletedID {
		t.Errorf("RoomDeleted has different room id, expect: %s, got: %s", deletedID, got)
	}
	if got := ev.Name; got != r.Name {
		t.Errorf("RoomDeleted has different room name, expect: %s, got: %s", r.Name, got)
	}
	if got, expect := len(ev.MemberIDs), len(r.MemberIDs()); got != expect {
		t.Errorf("RoomDeleted has dieffrent room members size, expect: %d, got: %d", expect, got)
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

	if !r.HasMember(u) {
		t.Errorf("AddMember does not add any member to the room")
	}

	// room has two events: Created, AddedMember.
	if got := len(r.Events()); got != 2 {
		t.Errorf("room has no event")
	}
	if _, ok := r.Events()[1].(RoomAddedMember); !ok {
		t.Errorf("invalid event is added")
	}
}
