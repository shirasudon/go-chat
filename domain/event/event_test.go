package event

import "testing"

func TestEventEmbd(t *testing.T) {
	for _, ev := range []struct {
		Name string
		Event
		ExpectType     Type
		ExpectStreamID StreamID
	}{
		{"EventEmbd", EventEmbd{}, TypeNone, NoneStream},
		{"UserEventEmbd", UserEventEmbd{}, TypeNone, UserStream},
		{"UserCreated", UserCreated{}, TypeUserCreated, UserStream},
		{"UserAddedFriend", UserAddedFriend{}, TypeUserAddedFriend, UserStream},
		{"RoomEventEmbd", RoomEventEmbd{}, TypeNone, RoomStream},
		{"RoomCreated", RoomCreated{}, TypeRoomCreated, RoomStream},
		{"RoomDeleted", RoomDeleted{}, TypeRoomDeleted, RoomStream},
		{"RoomAddedMember", RoomAddedMember{}, TypeRoomAddedMember, RoomStream},
		{"RoomMessagesReadByUser", RoomMessagesReadByUser{}, TypeRoomMessagesReadByUser, RoomStream},
		{"MessageEventEmbd", MessageEventEmbd{}, TypeNone, MessageStream},
		{"MessageCreated", MessageCreated{}, TypeMessageCreated, MessageStream},
		{"ActiveClientActivated", ActiveClientActivated{}, TypeActiveClientActivated, NoneStream},
		{"ActiveClientInactivated", ActiveClientInactivated{}, TypeActiveClientInactivated, NoneStream},
		{"ExternalEventEmbd", ExternalEventEmbd{}, TypeExternal, NoneStream},
	} {
		if ev.Type() != ev.ExpectType {
			t.Errorf("%v: different event type, got: %v, expect: %v", ev.Name, ev.Type(), ev.ExpectType)
		}
		if ev.StreamID() != ev.ExpectStreamID {
			t.Errorf("%v: different event stream id, got: %v, expect: %v", ev.Name, ev.StreamID(), ev.ExpectStreamID)
		}
	}
}

type NewEvent struct{ ExternalEventEmbd }

func (NewEvent) TypeString() string { return "new_event" }

func TestTypeString(t *testing.T) {
	ev := UserCreated{}
	if got, expect := TypeString(ev), ev.Type().String(); got != expect {
		t.Errorf("different type string, expect: %v, got: %v", expect, got)
	}

	if got, expect := TypeString(NewEvent{}), "new_event"; got != expect {
		t.Errorf("different type string, expect: %v, got: %v", expect, got)
	}

	if got, expect := TypeString(ExternalEventEmbd{}), "new_external_type"; got != expect {
		t.Errorf("different type string, expect: %v, got: %v", expect, got)
	}
}
