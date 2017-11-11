package inmemory

import (
	"context"
	"sort"

	"github.com/shirasudon/go-chat/chat"
	"github.com/shirasudon/go-chat/domain"
)

type RoomRepository struct {
	domain.EmptyTxBeginner
}

func NewRoomRepository() *RoomRepository {
	return &RoomRepository{}
}

var (
	DummyRoom1 = domain.Room{ID: 1, Name: "title1", MemberIDSet: domain.NewUserIDSet()}
	DummyRoom2 = domain.Room{ID: 2, Name: "title2", MemberIDSet: domain.NewUserIDSet(2, 3)}
	DummyRoom3 = domain.Room{ID: 3, Name: "title3", MemberIDSet: domain.NewUserIDSet(2)}

	roomMap = map[uint64]*domain.Room{
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

var roomCounter uint64 = uint64(len(roomMap))

func (repo *RoomRepository) FindAllByUserID(ctx context.Context, userID uint64) ([]domain.Room, error) {
	rooms := make([]domain.Room, 0, 4)
	for roomID, userIDs := range roomToUsersMap {
		if userIDs[userID] {
			rooms = append(rooms, *roomMap[roomID])
		}
	}
	sort.Slice(rooms, func(i, j int) bool { return rooms[i].ID < rooms[j].ID })
	return rooms, nil
}

func (repo *RoomRepository) Store(ctx context.Context, r domain.Room) (uint64, error) {
	if r.NotExist() {
		return repo.Create(ctx, r)
	} else {
		return repo.Update(ctx, r)
	}
}

func (repo *RoomRepository) Create(ctx context.Context, r domain.Room) (uint64, error) {
	roomCounter += 1
	r.ID = roomCounter
	roomMap[r.ID] = &r

	memberIDs := r.MemberIDs()
	userIDs := make(map[uint64]bool, len(memberIDs))
	for _, memberID := range memberIDs {
		userIDs[memberID] = true
	}
	roomToUsersMap[r.ID] = userIDs

	return r.ID, nil
}

func (repo *RoomRepository) Update(ctx context.Context, r domain.Room) (uint64, error) {
	if _, ok := roomMap[r.ID]; !ok {
		return 0, chat.NewInfraError("room(id=%d) is not in the datastore", r.ID)
	}

	// update room
	roomMap[r.ID] = &r

	userIDs := roomToUsersMap[r.ID]
	if userIDs == nil {
		userIDs = make(map[uint64]bool)
		roomToUsersMap[r.ID] = userIDs
	}

	// prepare user existance to off.
	for uid, _ := range userIDs {
		userIDs[uid] = false
	}
	// set user existance to on.
	for _, memberID := range r.MemberIDs() {
		userIDs[memberID] = true
	}
	// remove users deleteted from the room.
	for uid, exist := range userIDs {
		if !exist {
			delete(userIDs, uid)
		}
	}

	return r.ID, nil
}

func (repo *RoomRepository) Remove(ctx context.Context, r domain.Room) error {
	delete(roomMap, r.ID)
	delete(roomToUsersMap, r.ID)
	return nil
}

func (repo *RoomRepository) Find(ctx context.Context, roomID uint64) (domain.Room, error) {
	if room, ok := roomMap[roomID]; ok {
		return *room, nil
	}
	return domain.Room{}, ErrNotFound
}
