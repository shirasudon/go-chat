package domain

import (
	"context"
	"errors"
	"time"
)

type MessageRepository interface {
	TxBeginner

	Find(ctx context.Context, msgID uint64) (Message, error)

	// LatestRoomMessages returns latest n-messages from the room having room id.
	LatestRoomMessages(ctx context.Context, roomID uint64, n int) ([]Message, error)
	// PreviousRoomMessages returns n-messages before offset message.
	// offset message is excluded from result.
	// Typically offset message is got by first of the result of LatestRoomMessages().
	PreviousRoomMessages(ctx context.Context, offset Message, n int) ([]Message, error)

	// Store stores given message to the repository.
	// user need not to set ID for message since it is auto set
	// when message is newly.
	// It returns stored Message ID and error.
	Store(ctx context.Context, m Message) (uint64, error)
}

type Message struct {
	// ID and CreatedAt are auto set.
	ID        uint64    `db:"id"`
	CreatedAt time.Time `db:"created_at"`

	Content string `db:"content"`
	UserID  uint64 `db:"user_id"`
	RoomID  uint64 `db:"room_id"`
	Deleted bool   `db:"deleted"`
}

// NewMessage creates new message into the reposiotry.
// It returns created message, its event and some error.
func NewMessage(
	ctx context.Context,
	msgs MessageRepository,
	u User,
	r Room,
	content string,
) (Message, MessageCreated, error) {
	if u.IsNew() {
		return Message{}, MessageCreated{}, errors.New("the user not in the datastore, can not create new message")
	}
	if r.IsNew() {
		return Message{}, MessageCreated{}, errors.New("the room not in the datastore, can not create new message")
	}

	m := Message{
		ID:        0,
		CreatedAt: time.Now(),
		Content:   content,
		UserID:    u.ID,
		RoomID:    r.ID,
		Deleted:   false,
	}

	var err error
	m.ID, err = msgs.Store(ctx, m)
	if err != nil {
		return Message{}, MessageCreated{}, err
	}

	ev := MessageCreated{MessageID: m.ID}
	return m, ev, nil
}

func (m *Message) ReadBy(u User) (MessageReadByUser, error) {
	if u.IsNew() {
		return MessageReadByUser{}, errors.New("the user not in the datastore, can not read any message")
	}
	ev := MessageReadByUser{
		MessageID: m.ID,
		UserID:    u.ID,
	}
	return ev, nil
}

// Event for the message is created.
type MessageCreated struct {
	MessageID uint64
}

func (MessageCreated) EventType() EventType { return EventMessageCreated }

// Event for the message is read by the user.
type MessageReadByUser struct {
	MessageID uint64
	UserID    uint64
}

func (MessageReadByUser) EventType() EventType { return EventMessageReadByUser }
