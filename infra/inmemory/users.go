package inmemory

import (
	"context"
	"errors"
	"fmt"

	"github.com/shirasudon/go-chat/domain"
)

type UserRepository struct {
	domain.EmptyTxBeginner
}

var ErrNotFound = errors.New("user not found")

var (
	DummyUser  = domain.User{ID: 0, Name: "user", Password: "password"}
	DummyUser2 = domain.User{ID: 2, Name: "user2", Password: "password"}
	DummyUser3 = domain.User{ID: 3, Name: "user3", Password: "password"}

	userMap = map[uint64]domain.User{
		0: DummyUser,
		2: DummyUser2,
		3: DummyUser3,
	}

	userToUsersMap = map[uint64]map[uint64]bool{
		// user(id=2) relates with the user(id=3).
		2: {3: true},
	}
)

var userCounter uint64 = uint64(len(userMap))

func (repo UserRepository) FindByNameAndPassword(ctx context.Context, name, password string) (domain.User, error) {
	for _, u := range userMap {
		if name == u.Name && password == u.Password {
			return u, nil
		}
	}
	return domain.User{}, ErrNotFound
}

func (repo UserRepository) Store(ctx context.Context, u domain.User) (uint64, error) {
	if u.NotExist() {
		return repo.Create(ctx, u)
	} else {
		return repo.Update(ctx, u)
	}
}

func (repo *UserRepository) Create(ctx context.Context, u domain.User) (uint64, error) {
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
	if _, ok := userMap[u.ID]; !ok {
		return 0, fmt.Errorf("user(id=%d) is not in the datastore", u.ID)
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

func (repo UserRepository) ExistByNameAndPassword(ctx context.Context, name, password string) bool {
	_, err := repo.FindByNameAndPassword(ctx, name, password)
	return err == nil
}

func (repo UserRepository) Find(ctx context.Context, id uint64) (domain.User, error) {
	u, ok := userMap[id]
	if ok {
		return u, nil
	}
	return DummyUser, ErrNotFound
}

func (repo UserRepository) FindAllByUserID(ctx context.Context, id uint64) ([]domain.User, error) {
	userIDs, ok := userToUsersMap[id]
	if !ok || len(userIDs) == 0 {
		return nil, ErrNotFound
	}

	us := make([]domain.User, 0, len(userIDs))
	for id, _ := range userIDs {
		if u, ok := userMap[id]; ok {
			us = append(us, u)
		}
	}

	if len(us) == 0 {
		return nil, ErrNotFound
	}
	return us, nil
}
