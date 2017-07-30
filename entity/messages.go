package entity

import (
	"context"
	"time"
)

type Message struct {
	// ID and CreatedAt are auto set.
	ID        uint64    `db:"id"`
	CreatedAt time.Time `db:"created_at"`

	Content string `db:"content"`
	UserID  uint64 `db:"user_id"`
	RoomID  uint64 `db:"room_id"`
}

type MessageRepository interface {
	// LatestRoomMessages returns latest n-messages from the room having room id.
	LatestRoomMessages(ctx context.Context, roomID uint64, n int) ([]Message, error)
	// PreviousRoomMessages returns n-messages before offset message.
	// offset message is excluded from result.
	// Typically offset message is got by first of the result of LatestRoomMessages().
	PreviousRoomMessages(ctx context.Context, offset Message, n int) ([]Message, error)

	// Save stores given message to the reposiotry. user need to set ID and CreatedAt for
	// message since these are auto set.
	// It returns registered ID and error.
	Save(m Message) (uint64, error)
}

type MessageRepositoryStub struct {
	messages []Message
}

func NewMessageRepositoryStub() *MessageRepositoryStub {
	return &MessageRepositoryStub{
		messages: make([]Message, 0, 100),
	}
}

func (repo *MessageRepositoryStub) AddToRoom(roomID, m Message) error {
	repo.messages = append(repo.messages, m)
	return nil
}

func (repo *MessageRepositoryStub) GetFromRoom(roomID int64, n int) ([]Message, error) {
	if n > len(repo.messages) {
		n = len(repo.messages)
	}
	return repo.messages[:n], nil
}
