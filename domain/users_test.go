package domain

import (
	"testing"
)

func TestUserCreated(t *testing.T) {
	u, ev := NewUser("user", "password", NewUserIDSet(1))
	// check whether user has one event,
	events := u.Events()
	if got := len(events); got != 1 {
		t.Errorf("user has no event after UserCreated")
	}
	if _, ok := events[0].(UserCreated); !ok {
		t.Errorf("invalid event state for the user")
	}

	// check whether user created event is valid.
	if got := ev.Name; got != "user" {
		t.Errorf("UserCreated has different user name, expect: %s, got: %s", "user", got)
	}
	if got := len(ev.FriendIDs); got != 1 {
		t.Errorf("UseerCreated has dieffrent friends size, expect: %d, got: %d", 1, got)
	}

}

func TestUserAddFriendSuccess(t *testing.T) {
	u, _ := NewUser("user", "password", NewUserIDSet())
	u.ID = 1 // it may not be allowed at application side.
	friend := User{ID: u.ID + 1}
	ev, err := u.AddFriend(friend)
	if err != nil {
		t.Fatal(err)
	}
	if got := ev.UserID; got != u.ID {
		t.Errorf("UserAddedFriend has different user id, expect: %d, got: %d", u.ID, got)
	}
	if got := ev.AddedFriendID; got != friend.ID {
		t.Errorf("UserAddedFriend has different friend id, expect: %d, got: %d", friend.ID, got)
	}

	if !u.HasFriend(friend) {
		t.Errorf("AddFriend could not add friend to the user")
	}

	// user has two events: Created, AddedFriend.
	if got := len(u.Events()); got != 2 {
		t.Errorf("user has no event")
	}
	if _, ok := u.Events()[1].(UserAddedFriend); !ok {
		t.Errorf("invalid event is added")
	}
}

func TestUserAddFriendFail(t *testing.T) {
	// fail case: Add itself as friend.
	u, _ := NewUser("user", "password", NewUserIDSet())
	u.ID = 1 // it may not be allowed at application side.
	_, err := u.AddFriend(u)
	if err == nil {
		t.Fatal("add itself as friend but no error")
	}

	// user has one events: Created.
	if got := len(u.Events()); got != 1 {
		t.Errorf("user has invalid event state")
	}
}
