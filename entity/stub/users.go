package stub

import (
	"context"
	"errors"

	"github.com/shirasudon/go-chat/entity"
)

type UserRepository struct{}

var ErrNotFound = errors.New("user not found")

var (
	DummyUser  = entity.User{ID: 0, Name: "user", Password: "password"}
	DummyUser2 = entity.User{ID: 2, Name: "user2", Password: "password"}
	DummyUser3 = entity.User{ID: 3, Name: "user3", Password: "password"}

	userMap = map[uint64]entity.User{
		0: DummyUser,
		2: DummyUser2,
		3: DummyUser3,
	}

	userToUsersMap = map[uint64][]entity.User{
		2: {DummyUser3},
	}
)

func (repo UserRepository) FindByNameAndPassword(ctx context.Context, name, password string) (entity.User, error) {
	for _, u := range userMap {
		if name == u.Name && password == u.Password {
			return u, nil
		}
	}
	return entity.User{}, ErrNotFound
}

func (repo UserRepository) Store(ctx context.Context, u entity.User) (uint64, error) {
	panic("not implement")
}

func (repo UserRepository) ExistByNameAndPassword(ctx context.Context, name, password string) bool {
	_, err := repo.FindByNameAndPassword(ctx, name, password)
	return err == nil
}

func (repo UserRepository) Find(ctx context.Context, id uint64) (entity.User, error) {
	u, ok := userMap[id]
	if ok {
		return u, nil
	}
	return DummyUser, ErrNotFound
}

func (repo UserRepository) FindAllByUserID(ctx context.Context, id uint64) ([]entity.User, error) {
	us, ok := userToUsersMap[id]
	if ok {
		return us, nil
	}
	return nil, ErrNotFound
}
