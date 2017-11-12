package chat

import (
	"context"
	"time"

	"github.com/shirasudon/go-chat/domain"
)

//go:generate mockgen -destination=../internal/mocks/mock_queryer.go -package=mocks github.com/shirasudon/go-chat/chat UserQueryer,RoomQueryer,MessageQueryer

// Queryers is just data struct which have
// some XXXQueryers.
type Queryers struct {
	UserQueryer
	RoomQueryer
	MessageQueryer
}

// UserQueryer queries users stored in the data-store.
type UserQueryer interface {
	// Find a user specified by userID and return it.
	// It returns error if not found.
	Find(ctx context.Context, userID uint64) (domain.User, error)
}

// RoomQueryer queries rooms stored in the data-store.
type RoomQueryer interface {
	// Find a room specified by roomID and return it.
	// It returns error if not found.
	Find(ctx context.Context, roomID uint64) (domain.Room, error)

	// Find all rooms which user has.
	FindAllByUserID(ctx context.Context, userID uint64) ([]domain.Room, error)
}

// MessageQueryer queries messages stored in the data-store.
type MessageQueryer interface {

	// Find all messages from the room specified by room_id.
	// The returned messages are, ordered by latest created at,
	// all of before specified before time,
	// and the number of messages is limted to less than
	// specified limit.
	// It returns error if infrastructure raise some errors.
	FindRoomMessagesOrderByLatest(ctx context.Context, roomID uint64, before time.Time, limit int) ([]domain.Message, error)
}
