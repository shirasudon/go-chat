package domain

import (
	"context"
	"errors"
	"fmt"
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
	EventHolder

	// ID and CreatedAt are auto set.
	ID        uint64    `db:"id"`
	CreatedAt time.Time `db:"created_at"`

	Content string `db:"content"`
	UserID  uint64 `db:"user_id"`
	RoomID  uint64 `db:"room_id"`
	Deleted bool   `db:"deleted"`
}

// NewRoomMessage creates new message for the specified room.
// The created message is immediately stored into the repository.
// It returns new message holding event message created and error if any.
func NewRoomMessage(
	ctx context.Context,
	msgs MessageRepository,
	u User,
	r Room,
	content string,
) (Message, error) {
	if u.IsNew() {
		return Message{}, errors.New("the user not in the datastore, can not create new message")
	}
	if r.IsNew() {
		return Message{}, errors.New("the room not in the datastore, can not create new message")
	}
	if !r.HasMember(u) {
		return Message{}, fmt.Errorf("user(id=%d) not a member of the room(id=%d), can not create message", u.ID, r.ID)
	}

	m := Message{
		EventHolder: NewEventHolder(),
		ID:          0,
		CreatedAt:   time.Now(),
		Content:     content,
		UserID:      u.ID,
		RoomID:      r.ID,
		Deleted:     false,
	}
	id, err := msgs.Store(ctx, m)
	if err != nil {
		return Message{}, err
	}
	m.ID = id

	ev := MessageCreated{
		MessageID:  m.ID,
		SenderName: u.Name,
		Content:    content,
	}
	m.AddEvent(ev)

	return m, nil
}

// ReadBy marks the message to read by specified user.
// It returns such event, which is contained the message,
// and error if any.
func (m *Message) ReadBy(u User) (MessageReadByUser, error) {
	if u.IsNew() {
		return MessageReadByUser{}, errors.New("the user not in the datastore, can not read any message")
	}
	ev := MessageReadByUser{
		MessageID: m.ID,
		UserID:    u.ID,
	}
	m.AddEvent(ev)
	return ev, nil
}

// Event for the message is created.
type MessageCreated struct {
	MessageID  uint64
	SenderName string
	Content    string
}

func (MessageCreated) EventType() EventType { return EventMessageCreated }

// Event for the message is read by the user.
type MessageReadByUser struct {
	MessageID uint64
	UserID    uint64
}

func (MessageReadByUser) EventType() EventType { return EventMessageReadByUser }
