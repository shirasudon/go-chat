package inmemory

import (
	"context"
	"sort"
	"sync"

	"github.com/shirasudon/go-chat/chat"
	"github.com/shirasudon/go-chat/chat/queried"
	"github.com/shirasudon/go-chat/domain"
)

type UserRepository struct {
	domain.EmptyTxBeginner
}

var (
	DummyUser = domain.User{
		ID:        0,
		Name:      "user",
		FirstName: "u-",
		LastName:  "ser",
		Password:  "password",
	}
	DummyUser2 = domain.User{
		ID:        2,
		Name:      "user2",
		FirstName: "u-",
		LastName:  "ser",
		Password:  "password",
		FriendIDs: domain.NewUserIDSet(3),
	}
	DummyUser3 = domain.User{
		ID:        3,
		Name:      "user3",
		FirstName: "u-",
		LastName:  "ser",
		Password:  "password",
	}

	userMapMu *sync.RWMutex = new(sync.RWMutex)

	userMap = map[uint64]domain.User{
		0: DummyUser,
		2: DummyUser2,
		3: DummyUser3,
	}

	userNameUniqueMap = map[string]bool{
		DummyUser.Name:  true,
		DummyUser2.Name: true,
		DummyUser3.Name: true,
	}

	userToUsersMap = map[uint64]map[uint64]bool{
		// user(id=2) relates with the user(id=3).
		2: {3: true},
	}
)

func errUserNotFound(userID uint64) *chat.NotFoundError {
	return chat.NewNotFoundError("user (id=%v) is not found")
}

var userCounter uint64 = uint64(len(userMap))

func (repo UserRepository) Store(ctx context.Context, u domain.User) (uint64, error) {
	if u.NotExist() {
		return repo.Create(ctx, u)
	} else {
		return repo.Update(ctx, u)
	}
}

func (repo *UserRepository) Create(ctx context.Context, u domain.User) (uint64, error) {
	userMapMu.Lock()
	defer userMapMu.Unlock()

	if exist := userNameUniqueMap[u.Name]; exist {
		return 0, chat.NewInfraError("user name(%v) already exist and not allowed", u.Name)
	}
	userNameUniqueMap[u.Name] = true

	userCounter += 1
	u.ID = roomCounter
	userMap[u.ID] = u

	friendIDs := u.FriendIDs.List()
	userIDs := make(map[uint64]bool, len(friendIDs))
	for _, friendID := range friendIDs {
		userIDs[friendID] = true
	}
	userToUsersMap[u.ID] = userIDs

	return u.ID, nil
}

func (repo *UserRepository) Update(ctx context.Context, u domain.User) (uint64, error) {
	userMapMu.Lock()
	defer userMapMu.Unlock()

	if _, ok := userMap[u.ID]; !ok {
		return 0, chat.NewInfraError("user(id=%d) is not in the datastore", u.ID)
	}

	// update user
	userMap[u.ID] = u

	userIDs := userToUsersMap[u.ID]
	if userIDs == nil {
		userIDs = make(map[uint64]bool)
		userToUsersMap[u.ID] = userIDs
	}

	// prepare user existance to off.
	for uid, _ := range userIDs {
		userIDs[uid] = false
	}
	// set user existance to on.
	for _, friendID := range u.FriendIDs.List() {
		userIDs[friendID] = true
	}
	// remove users deleteted from the room.
	for uid, exist := range userIDs {
		if !exist {
			delete(userIDs, uid)
		}
	}

	return u.ID, nil
}

func (repo UserRepository) Find(ctx context.Context, id uint64) (domain.User, error) {
	userMapMu.RLock()
	defer userMapMu.RUnlock()

	u, ok := userMap[id]
	if ok {
		return u, nil
	}
	return DummyUser, errUserNotFound(id)
}

func (repo UserRepository) FindAllByUserID(ctx context.Context, id uint64) ([]domain.User, error) {
	userMapMu.RLock()
	defer userMapMu.RUnlock()

	userIDs, ok := userToUsersMap[id]
	if !ok || len(userIDs) == 0 {
		return nil, errUserNotFound(id)
	}

	us := make([]domain.User, 0, len(userIDs))
	for id, _ := range userIDs {
		if u, ok := userMap[id]; ok {
			us = append(us, u)
		}
	}

	if len(us) == 0 {
		return nil, chat.NewNotFoundError("any user id (%v) is not found", userIDs)
	}
	sort.Slice(us, func(i, j int) bool { return us[i].ID < us[j].ID })
	return us, nil
}

func (repo UserRepository) FindByNameAndPassword(ctx context.Context, name, password string) (*queried.AuthUser, error) {
	userMapMu.RLock()
	defer userMapMu.RUnlock()

	for _, u := range userMap {
		if name == u.Name && password == u.Password {
			return &queried.AuthUser{
				ID:       u.ID,
				Name:     u.Name,
				Password: u.Password,
			}, nil
		}
	}
	return nil, ErrNotFound
}

func createUserProfile(u *domain.User) queried.UserProfile {
	return queried.UserProfile{
		UserID:    u.ID,
		UserName:  u.Name,
		FirstName: u.FirstName,
		LastName:  u.LastName,
	}
}

func (repo UserRepository) FindUserRelation(ctx context.Context, userID uint64) (*queried.UserRelation, error) {
	// TODO: run constructing service by using event,
	// then just return already constructed value.
	userMapMu.RLock()

	user, ok := userMap[userID]
	if !ok {
		userMapMu.RUnlock()
		return nil, errUserNotFound(userID)
	}

	friends := make([]queried.UserProfile, 0, 4)
	for _, id := range user.FriendIDs.List() {
		if friend, ok := userMap[id]; ok {
			friends = append(friends, createUserProfile(&friend))
		}
	}

	userMapMu.RUnlock()

	roomMapMu.RLock()

	rooms := make([]queried.UserRoom, 0, 4)
	for rID, userIDs := range roomToUsersMap {
		if _, ok := userIDs[userID]; ok {
			r := roomMap[rID]
			rooms = append(rooms, queried.UserRoom{
				RoomID:   rID,
				RoomName: r.Name,
			})
		}
	}

	roomMapMu.RUnlock()

	return &queried.UserRelation{
		UserProfile: createUserProfile(&user),
		Friends:     friends,
		Rooms:       rooms,
	}, nil
}
