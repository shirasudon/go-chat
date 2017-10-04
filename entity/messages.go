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
	Deleted bool   `db:"deleted"`
}

type MessageRepository interface {
	TxBeginner

	// LatestRoomMessages returns latest n-messages from the room having room id.
	LatestRoomMessages(ctx context.Context, roomID uint64, n int) ([]Message, error)
	// PreviousRoomMessages returns n-messages before offset message.
	// offset message is excluded from result.
	// Typically offset message is got by first of the result of LatestRoomMessages().
	PreviousRoomMessages(ctx context.Context, offset Message, n int) ([]Message, error)

	// Save stores given message to the reposiotry.
	// user need not to set ID and CreatedAt for
	// message since these are auto set.
	// It returns stored Message ID and error.
	Add(ctx context.Context, m Message) (uint64, error)

	// ReadMessage marks the messages specified by roomID, MessageIDs
	// to be read by user specified by userID.
	// It returns some error for
	ReadMessage(ctx context.Context, roomID, userID uint64, messageIDs []uint64) error
}
