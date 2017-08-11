package sqlite3

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/shirasudon/go-chat/entity"
)

type RoomRepository struct {
	db *sqlx.DB

	users entity.UserRepository
}

func newRoomRepository(db *sqlx.DB) (*RoomRepository, error) {
	return &RoomRepository{
		db: db,
	}, nil
}

func (repo RoomRepository) GetUserRooms(ctx context.Context, userID uint64) ([]entity.Room, error) {
	panic("not implemented")
}

func (repo RoomRepository) Add(ctx context.Context, r entity.Room) (entity.Room, error) {
	panic("not implemented")
}

func (repo RoomRepository) Remove(ctx context.Context, r entity.Room) error {
	panic("not implemented")
}

func (repo RoomRepository) RoomHasMember(ctx context.Context, roomID uint64, userID uint64) bool {
	panic("not implemented")
}
