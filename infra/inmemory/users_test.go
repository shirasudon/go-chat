package inmemory

import (
	"context"
	"testing"
)

func TestFindUserRelation(t *testing.T) {
	repo := &UserRepository{}

	// case success
	const TestUserID = uint64(2)

	user, err := repo.Find(context.Background(), TestUserID)
	if err != nil {
		t.Fatal(err)
	}

	relation, err := repo.FindUserRelation(context.Background(), TestUserID)
	if err != nil {
		t.Fatal(err)
	}

	if expect, got := TestUserID, relation.UserID; expect != got {
		t.Errorf("different user id, expect: %v, got: %v", expect, got)
	}
	if expect, got := user.Name, relation.UserName; expect != got {
		t.Errorf("different user name, expect: %v, got: %v", expect, got)
	}
	if expect, got := user.FirstName, relation.FirstName; expect != got {
		t.Errorf("different user first name, expect: %v, got: %v", expect, got)
	}
	if expect, got := user.LastName, relation.LastName; expect != got {
		t.Errorf("different user last name, expect: %v, got: %v", expect, got)
	}

	if expect, got := 1, len(relation.Friends); expect != got {
		t.Errorf("different number of friends, expect: %v, got: %v", expect, got)
	}
	if expect, got := 2, len(relation.Rooms); expect != got {
		t.Errorf("different number of rooms, expect: %v, got: %v", expect, got)
	}

	// case fail
	const NotExistUserID = uint64(99)
	if _, err := repo.FindUserRelation(context.Background(), NotExistUserID); err == nil {
		t.Fatalf("query no exist user ID (%v) but no error", NotExistUserID)
	}
}
