package entity

import (
	"testing"
)

func TestRoomAddMember(t *testing.T) {
	r := Room{}
	err := r.AddMember(1)
	if err != nil {
		t.Fatal(err)
	}
	if !r.HasMember(1) {
		t.Errorf("AddMember does not add any member to the room")
	}
}
