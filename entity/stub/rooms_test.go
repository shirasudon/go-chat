package stub

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
