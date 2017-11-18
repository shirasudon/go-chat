package domain

import (
	"context"
	"errors"
	"fmt"
	"time"
)

//go:generate mockgen -destination=../internal/mocks/mock_messages.go -package=mocks github.com/shirasudon/go-chat/domain MessageRepository

type MessageRepository interface {
	TxBeginner

	Find(ctx context.Context, msgID uint64) (Message, error)

	// Store stores given message to the repository.
	// user need not to set ID for message since it is auto set
	// when message is newly.
	// It returns stored Message ID and error.
	Store(ctx context.Context, m Message) (uint64, error)

	// RemoveAllByRoomID removes all messages related with roomID.
	RemoveAllByRoomID(ctx context.Context, roomID uint64) error
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
	if u.NotExist() {
		return Message{}, errors.New("the user not in the datastore, can not create new message")
	}
	if r.NotExist() {
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
		MessageID: m.ID,
		RoomID:    m.RoomID,
		CreatedBy: u.ID,
		Content:   content,
	}
	ev.Occurs()
	m.AddEvent(ev)

	return m, nil
}

func (m *Message) NotExist() bool {
	return m == nil || m.ID == 0
}

// ReadBy marks the message to read by specified user.
// It returns such event, which is contained the message,
// and error if any.
func (m *Message) ReadBy(u User) (MessageReadByUser, error) {
	if u.NotExist() {
		return MessageReadByUser{}, errors.New("the user not in the datastore, can not read any message")
	}
	ev := MessageReadByUser{
		MessageID: m.ID,
		UserID:    u.ID,
	}
	ev.Occurs()
	m.AddEvent(ev)
	return ev, nil
}

// -----------------------
// Message events
// -----------------------

// Event for the message is created.
type MessageCreated struct {
	EventEmbd
	MessageID uint64 `json:"message_id"`
	RoomID    uint64 `json:"room_id"`
	CreatedBy uint64 `json:"created_by"`
	Content   string `json:"content"`
}

func (MessageCreated) EventType() EventType { return EventMessageCreated }

// Event for the message is read by the user.
type MessageReadByUser struct {
	EventEmbd
	MessageID uint64 `json:"message_id"`
	UserID    uint64 `json:"user_id"`
}

func (MessageReadByUser) EventType() EventType { return EventMessageReadByUser }
