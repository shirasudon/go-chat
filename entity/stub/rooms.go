package stub

import (
	"context"

	"github.com/shirasudon/go-chat/entity"
)

type RoomRepository struct{}

var (
	DummyRoom  = entity.Room{ID: 0, Name: "title"}
	DummyRoom2 = entity.Room{ID: 2, Name: "title2"}
	DummyRoom3 = entity.Room{ID: 3, Name: "title3"}

	DummyRooms = []entity.Room{
		DummyRoom2,
		DummyRoom3,
	}
)

func (repo *RoomRepository) GetUserRooms(ctx context.Context, userID uint64) ([]entity.Room, error) {
	return DummyRooms, nil
}

func (repo *RoomRepository) Add(ctx context.Context, r entity.Room) (entity.Room, error) {
	return r, nil
}

func (repo *RoomRepository) Remove(ctx context.Context, r entity.Room) error {
	return nil
}

func (repo *RoomRepository) RoomHasMember(ctx context.Context, roomID uint64, userID uint64) bool {
	return true
}
