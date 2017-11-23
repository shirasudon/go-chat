package inmemory

import (
	"context"
	"testing"
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

	memberExists := false
	for _, m := range info.Members {
		if m.UserID == TestUserID {
			memberExists = true
		}
	}
	if !memberExists {
		t.Errorf("query parameter is not found in the result, user id %v", TestUserID)
	}

	// case fail
	if _, err := repo.FindRoomInfo(context.Background(), NotExistUserID, TestRoomID); err == nil {
		t.Fatalf("query no exist user ID (%v) but no error", NotExistUserID)
	}
	if _, err := repo.FindRoomInfo(context.Background(), TestUserID, NotExistRoomID); err == nil {
		t.Fatalf("query no exist room ID (%v) but no error", NotExistRoomID)
	}
}
