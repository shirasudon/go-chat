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
		Room: DummyRoom2,
		Members: map[uint64]entity.User{
			2: DummyUser2,
			3: DummyUser3,
		},
	}

	DummyRooms = []entity.Room{
		DummyRoom2,
		DummyRoom3,
	}
)

func (repo *RoomRepository) GetUserRooms(ctx context.Context, userID uint64) ([]entity.Room, error) {
	return DummyRooms, nil
}

func (repo *RoomRepository) Add(ctx context.Context, r entity.Room) (uint64, error) {
	return 1, nil
}

func (repo *RoomRepository) Remove(ctx context.Context, r entity.Room) error {
	return nil
}

func (repo *RoomRepository) Find(ctx context.Context, roomID uint64) (entity.Room, error) {
	return DummyRoom, nil
}

type RoomRelationRepository struct{}

func (repo *RoomRelationRepository) Find(ctx context.Context, roomID uint64) (entity.RoomRelation, error) {
	relation := DummyRoomRelation
	relation.ID = roomID
	return relation, nil
}

func (repo *RoomRelationRepository) ExistRoomMember(ctx context.Context, roomID uint64, userID uint64) bool {
	return true
}
