package stub

import (
	"context"
	"errors"

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

	userToUsersMap = map[uint64][]domain.User{
		2: {DummyUser3},
	}
)

func (repo UserRepository) FindByNameAndPassword(ctx context.Context, name, password string) (domain.User, error) {
	for _, u := range userMap {
		if name == u.Name && password == u.Password {
			return u, nil
		}
	}
	return domain.User{}, ErrNotFound
}

func (repo UserRepository) Store(ctx context.Context, u domain.User) (uint64, error) {
	panic("not implement")
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
	us, ok := userToUsersMap[id]
	if ok {
		return us, nil
	}
	return nil, ErrNotFound
}
