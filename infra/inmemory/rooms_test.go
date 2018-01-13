package inmemory

import (
	"context"
	"testing"
	"time"

	"github.com/shirasudon/go-chat/domain"
	"github.com/shirasudon/go-chat/domain/event"
)

func TestRoomVars(t *testing.T) {
	testUserID := uint64(2)
	repo := RoomRepository{}
	rooms, err := repo.FindAllByUserID(context.Background(), testUserID)
	if err != nil {
		t.Fatal(err)
	}
	if len(rooms) == 0 {
		t.Fatalf("user(id=%d) has no rooms", testUserID)
	}

	memberIDs := rooms[0].MemberIDs()
	if got := len(memberIDs); got != 2 {
		t.Errorf(
			"different member size for the room(%#v), expect %d, got %d",
			rooms[0], 2, got,
		)
	}
}

func TestFindRoomInfo(t *testing.T) {
	repo := &RoomRepository{}

	// case success
	const (
		TestUserID = uint64(2)
		TestRoomID = uint64(2)

		NotExistUserID = uint64(99)
		NotExistRoomID = uint64(99)
	)

	var (
		TimeNow = time.Now()
	)

	// setup read time for test user.
	{
		readTimes := &roomMap[TestRoomID].MemberReadTimes
		prevTime, _ := readTimes.Get(TestUserID)
		readTimes.Set(TestUserID, TimeNow)
		defer func() {
			readTimes.Set(TestUserID, prevTime)
		}()
	}

	info, err := repo.FindRoomInfo(context.Background(), TestUserID, TestRoomID)
	if err != nil {
		t.Fatal(err)
	}
	room := roomMap[2]

	if expect, got := room.ID, info.RoomID; expect != got {
		t.Errorf("different room id, expect: %v, got: %v", expect, got)
	}
	if expect, got := room.Name, info.RoomName; expect != got {
		t.Errorf("different room name, expect: %v, got: %v", expect, got)
	}
	if expect, got := len(info.Members), info.MembersSize; expect != got {
		t.Errorf("different number of members, expect: %v, got: %v", expect, got)
	}

	var (
		memberExists  = false
		memberReadNow = false
	)
	for _, m := range info.Members {
		if m.UserID == TestUserID {
			memberExists = true
			if m.MessageReadAt.Equal(TimeNow) {
				memberReadNow = true
			}
		}
	}
	if !memberExists {
		t.Errorf("query parameter is not found in the result, user id %v", TestUserID)
	}
	if !memberReadNow {
		t.Errorf("room member (id=%v) has no read time", TestUserID)
	}

	// case fail
	if _, err := repo.FindRoomInfo(context.Background(), NotExistUserID, TestRoomID); err == nil {
		t.Fatalf("query no exist user ID (%v) but no error", NotExistUserID)
	}
	if _, err := repo.FindRoomInfo(context.Background(), TestUserID, NotExistRoomID); err == nil {
		t.Fatalf("query no exist room ID (%v) but no error", NotExistRoomID)
	}
}

func TestRoomStore(t *testing.T) {
	repo := &RoomRepository{}

	const (
		FirstName  = "room1"
		SecondName = "room2"
	)

	var (
		newR = domain.Room{Name: FirstName}
		err  error
	)
	newR.AddEvent(event.RoomCreated{})
	// create
	newR.ID, err = repo.Store(context.Background(), newR)
	if err != nil {
		t.Fatal(err)
	}

	storedR, ok := roomMap[newR.ID]
	if !ok {
		t.Fatal("nothing room after Store: create")
	}
	if len(storedR.Events()) != 0 {
		t.Fatal("event should never be persited")
	}
	if storedR.Name != FirstName {
		t.Errorf("different stored room name, expect: %v, got: %v", FirstName, storedR.Name)
	}

	newR.Name = SecondName
	newR.AddEvent(event.RoomCreated{})
	// update
	newR.ID, err = repo.Store(context.Background(), newR)
	if err != nil {
		t.Fatal(err)
	}

	storedR, ok = roomMap[newR.ID]
	if !ok {
		t.Fatal("nothing room after Store: update")
	}
	if len(storedR.Events()) != 0 {
		t.Fatal("event should never be persited")
	}
	if storedR.Name != SecondName {
		t.Errorf("different stored room name, expect: %v, got: %v", SecondName, storedR.Name)
	}
}
