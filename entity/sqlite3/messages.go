package sqlite3

import (
	"context"

	"github.com/mzki/go-chat/entity"
)

type MessageRepository struct{}

func (mr MessageRepository) LatestRoomMessages(ctx context.Context, roomID uint64, n int) ([]entity.Message, error) {
	panic("not implemented")
}

func (mr MessageRepository) PreviousRoomMessages(ctx context.Context, offset entity.Message, n int) ([]entity.Message, error) {
	panic("not implemented")
}

func (mr MessageRepository) Save(m entity.Message) (uint64, error) {
	panic("not implemented")
}
