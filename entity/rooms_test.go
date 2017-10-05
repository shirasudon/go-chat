package entity

import (
	"testing"
)

func TestRoomAddMember(t *testing.T) {
	r := NewRoom(0, "test", make(map[uint64]bool))
	u := User{ID: 1}
	ev, err := r.AddMember(u)
	if err != nil {
		t.Fatal(err)
	}
	if got := ev.RoomID; got != r.ID {
		t.Errorf("RoomAddedMember has dieffrent room id, expect: %d, got: %d", r.ID, got)
	}
	if got := ev.AddedUserID; got != u.ID {
		t.Errorf("RoomAddedMember has dieffrent user id, expect: %d, got: %d", u.ID, got)
	}

	if !r.HasMember(u) {
		t.Errorf("AddMember does not add any member to the room")
	}
}

func TestRoomPostMessage(t *testing.T) {
	const TestContent = "content"
	r := NewRoom(0, "test", make(map[uint64]bool))
	u := User{ID: 1}
	_, _ = r.AddMember(u)

	// case success
	ev, err := r.PostMessage(u, TestContent)
	if err != nil {
		t.Fatal(err)
	}
	if got := ev.PostedRoomID; got != r.ID {
		t.Errorf("RoomPostedMessage has different room id, expect: %d, got: %d", r.ID, got)
	}
	if got := ev.PostUserID; got != u.ID {
		t.Errorf("RoomPostedMessage has different user id, expect: %d, got: %d", u.ID, got)
	}
	if got := ev.Content; got != TestContent {
		t.Errorf("RoomPostedMessage has different message content, expect: %d, got: %d", TestContent, got)
	}

	// case fail: post message by non-exist user
	nobody := User{ID: 0}
	if _, err := r.PostMessage(nobody, TestContent); err == nil {
		t.Error("post message by non-exist member, but success to post new message")
	}
}
