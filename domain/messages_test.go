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

func TestMessageCreated(t *testing.T) {
	ctx := context.Background()
	m, ev, err := NewMessage(ctx, msgRepo, User{ID: 1}, Room{ID: 1}, "content")
	if err != nil {
		t.Fatal(err)
	}
	// check whether message created event is valid.
	if got := ev.MessageID; got != m.ID {
		t.Errorf("MessageCreated has different messageID, expect: %v, got: %v", m.ID, got)
	}
}

func TestMessageReadByUser(t *testing.T) {
	ctx := context.Background()
	m, _, err := NewMessage(ctx, msgRepo, User{ID: 1}, Room{ID: 1}, "content")
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
