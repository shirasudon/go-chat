package domain

import (
	"context"
	"database/sql"
	"testing"
)

type MessageRepositoryStub struct{}

func (m *MessageRepositoryStub) BeginTx(context.Context, *sql.TxOptions) (Tx, error) {
	panic("not implemented")
}

func (m *MessageRepositoryStub) Find(ctx context.Context, msgID uint64) (Message, error) {
	panic("not implemented")
}

func (m *MessageRepositoryStub) LatestRoomMessages(ctx context.Context, roomID uint64, n int) ([]Message, error) {
	panic("not implemented")
}

func (m *MessageRepositoryStub) PreviousRoomMessages(ctx context.Context, offset Message, n int) ([]Message, error) {
	panic("not implemented")
}

func (m *MessageRepositoryStub) Store(ctx context.Context, msg Message) (uint64, error) {
	return msg.ID, nil
}

var msgRepo MessageRepository = &MessageRepositoryStub{}

func TestMessageCreatedSuccess(t *testing.T) {
	var (
		ctx  = context.Background()
		user = User{ID: 1}
		room = Room{ID: 1}
	)
	room.AddMember(user)
	m, ev, err := NewRoomMessage(ctx, msgRepo, user, room, "content")
	if err != nil {
		t.Fatal(err)
	}
	// check whether message created event is valid.
	if got := ev.MessageID; got != m.ID {
		t.Errorf("MessageCreated has different messageID, expect: %v, got: %v", m.ID, got)
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
		_, _, err := NewRoomMessage(ctx, msgRepo, user, room, "content")
		if err == nil {
			t.Errorf("invalid combination of user and room, but no error: user(%d), room(%d)", user.ID, room.ID)
		}
	}
}

func TestMessageReadByUser(t *testing.T) {
	ctx := context.Background()
	user, room := User{ID: 1}, Room{ID: 1}
	room.AddMember(user)
	m, _, err := NewRoomMessage(ctx, msgRepo, user, room, "content")
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
}
