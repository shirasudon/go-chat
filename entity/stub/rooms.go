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

	DummyRoomRelation = entity.RoomRelation{
		Members: []entity.User{DummyUser2, DummyUser3},
	}

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

func (repo *RoomRepository) Find(ctx context.Context, roomID uint64) (entity.Room, error) {
	return DummyRoom, nil
}

func (repo *RoomRepository) FindWithRelation(ctx context.Context, roomID uint64) (entity.Room, entity.RoomRelation, error) {
	relation := DummyRoomRelation
	relation.RoomID = roomID
	return DummyRoom, relation, nil
}

func (repo *RoomRepository) RoomHasMember(ctx context.Context, roomID uint64, userID uint64) bool {
	return true
}
