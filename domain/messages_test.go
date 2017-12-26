package domain

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/shirasudon/go-chat/domain/event"
)

type MessageRepositoryStub struct{}

func (m *MessageRepositoryStub) BeginTx(context.Context, *sql.TxOptions) (Tx, error) {
	panic("not implemented")
}

func (m *MessageRepositoryStub) Find(ctx context.Context, msgID uint64) (Message, error) {
	panic("not implemented")
}

func (m *MessageRepositoryStub) FindAllByRoomIDOrderByLatest(ctx context.Context, roomID uint64, n int) ([]Message, error) {
	panic("not implemented")
}

func (m *MessageRepositoryStub) FindPreviousMessagesOrderByLatest(ctx context.Context, offset Message, n int) ([]Message, error) {
	panic("not implemented")
}

func (m *MessageRepositoryStub) Store(ctx context.Context, msg Message) (uint64, error) {
	return msg.ID + 1, nil
}

func (m *MessageRepositoryStub) RemoveAllByRoomID(ctx context.Context, roomID uint64) error {
	panic("not implemented")
}

var msgRepo MessageRepository = &MessageRepositoryStub{}

func TestMessageCreatedSuccess(t *testing.T) {
	var (
		ctx  = context.Background()
		user = User{ID: 1}
		room = Room{ID: 1}
	)
	room.AddMember(user)
	m, err := NewRoomMessage(ctx, msgRepo, user, room, "content")
	if err != nil {
		t.Fatal(err)
	}

	// check whether message has valid ID
	if m.ID == 0 {
		t.Fatalf("message is created but has invalid ID(%d)", m.ID)
	}

	// check whether message created event is valid.
	events := m.Events()
	if len(events) != 1 {
		t.Fatalf("Message is created but message has no event for that.")
	}
	ev, ok := events[0].(event.MessageCreated)
	if !ok {
		t.Fatalf("Message is created but event is not a MessageCreated, got: %v", events[0])
	}
	if got := ev.MessageID; got != m.ID {
		t.Errorf("MessageCreated has different messageID, expect: %v, got: %v", m.ID, got)
	}
	if got := ev.Timestamp(); got == (time.Time{}) {
		t.Error("MessageCreated has no timestamp")
	}
}

func TestMessageCreatedFail(t *testing.T) {
	var (
		ctx = context.Background()
	)

	for _, testcase := range []struct {
		User
		Room
	}{
		{User{ID: 0}, Room{ID: 1}},
		{User{ID: 1}, Room{ID: 0}},
		{User{ID: 0}, Room{ID: 0}},
	} {
		user, room := testcase.User, testcase.Room
		room.AddMember(user)
		_, err := NewRoomMessage(ctx, msgRepo, user, room, "content")
		if err == nil {
			t.Errorf("invalid combination of user and room, but no error: user(%d), room(%d)", user.ID, room.ID)
		}
	}
}

func TestMessageReadByUser(t *testing.T) {
	ctx := context.Background()
	user, room := User{ID: 1}, Room{ID: 1}
	room.AddMember(user)
	m, err := NewRoomMessage(ctx, msgRepo, user, room, "content")
	if err != nil {
		t.Fatal(err)
	}

	u := User{ID: 1}
	ev, err := m.ReadBy(u)
	if err != nil {
		t.Fatal(err)
	}
	if got := ev.MessageID; got != m.ID {
		t.Errorf("MessageReadByUser has different message id, expect: %d, got: %d", m.ID, got)
	}
	if got := ev.UserID; got != u.ID {
		t.Errorf("MessageReadByUser has different user id, expect: %d, got: %d", u.ID, got)
	}
	if got := ev.RoomID; got != room.ID {
		t.Errorf("MessageReadByUser has different room id, expect: %d, got: %d", room.ID, got)
	}
	if got := ev.ReadAt; got == (time.Time{}) {
		t.Error("MessageReadByUser has no readAt time, expect time.Now()")
	}
	if got := ev.Timestamp(); got == (time.Time{}) {
		t.Error("MessageReadByUser has no timestamp")
	}

	// message holds 2 events: MessageCreated, MessageReadByUser
	events := m.Events()
	if len(events) != 2 {
		t.Fatalf("Message holds different events size, expect: %v, got: %v", 2, len(events))
	}
	ev2, ok := events[1].(event.MessageReadByUser)
	if !ok {
		t.Errorf("Message is read but event is not MessageReadByUser, got: %v", ev2)
	}
	if ev != ev2 {
		t.Errorf("returned event is not same as the holding event")
	}
	if got := ev2.Timestamp(); got == (time.Time{}) {
		t.Error("MessageReadByUser has no timestamp")
	}

	// fail case: not exist
	if _, err := m.ReadBy(User{ID: 0}); err == nil {
		t.Error("the message read by not exist user, but returns no error")
	}
}
