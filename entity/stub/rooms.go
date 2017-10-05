package stub

import (
	"context"

	"github.com/shirasudon/go-chat/entity"
)

type RoomRepository struct {
	entity.EmptyTxBeginner
}

var (
	DummyRoom1 = entity.NewRoom(1, "title1", make(map[uint64]bool))
	DummyRoom2 = entity.NewRoom(2, "title2", make(map[uint64]bool))
	DummyRoom3 = entity.NewRoom(3, "title3", make(map[uint64]bool))

	roomMap = map[uint64]*entity.Room{
		1: &DummyRoom1,
		2: &DummyRoom2,
		3: &DummyRoom3,
	}

	// Many-to-many mapping for Room-to-User.
	roomToUsersMap = map[uint64]map[uint64]bool{
		// room id = 2 has,
		2: {
			// user id = 2 and id = 3.
			2: true,
			3: true,
		},

		// room id = 3 has,
		3: {
			// user id = 2.
			2: true,
		},
	}
)

func init() {
	// initialize room-to-member relation.
	for _, room := range roomMap {
		for memberID, _ := range roomToUsersMap[room.ID] {
			if _, err := room.AddMember(entity.User{ID: memberID}); err != nil {
				panic(err)
			}
		}
	}
}

var roomCounter uint64 = uint64(len(roomMap))

func (repo *RoomRepository) FindAllByUserID(ctx context.Context, userID uint64) ([]entity.Room, error) {
	rooms := make([]entity.Room, 0, 4)
	for roomID, userIDs := range roomToUsersMap {
		if userIDs[userID] {
			rooms = append(rooms, *roomMap[roomID])
		}
	}
	return rooms, nil
}

func (repo *RoomRepository) Store(ctx context.Context, r entity.Room) (uint64, error) {
	roomCounter += 1
	r.ID = roomCounter
	roomMap[r.ID] = &r
	return r.ID, nil
}

func (repo *RoomRepository) Remove(ctx context.Context, roomID uint64) error {
	delete(roomMap, roomID)
	return nil
}

func (repo *RoomRepository) Find(ctx context.Context, roomID uint64) (entity.Room, error) {
	if room, ok := roomMap[roomID]; ok {
		return *room, nil
	}
	return entity.Room{}, ErrNotFound
}
