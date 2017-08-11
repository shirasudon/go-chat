package sqlite3

import (
	"context"

	"github.com/shirasudon/go-chat/entity"
)

type MessageRepository struct{}

func (mr MessageRepository) LatestRoomMessages(ctx context.Context, roomID uint64, n int) ([]entity.Message, error) {
	panic("not implemented")
}

func (mr MessageRepository) PreviousRoomMessages(ctx context.Context, offset entity.Message, n int) ([]entity.Message, error) {
	panic("not implemented")
}

func (mr MessageRepository) Add(ctx context.Context, m entity.Message) (entity.Message, error) {
	panic("not implemented")
}

func (mr MessageRepository) ReadMessage(ctx context.Context, roomID, userID uint64, messageIDs []uint64) error {
	panic("not implemented")
}
