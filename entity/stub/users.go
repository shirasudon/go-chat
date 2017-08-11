package stub

import (
	"context"
	"errors"

	"github.com/shirasudon/go-chat/entity"
)

type UserRepository struct{}

var ErrNotFound = errors.New("user not found")

var DummyUser = entity.User{ID: 0, Name: "user", Password: "password"}

func (repo UserRepository) Get(name string, password string) (entity.User, error) {
	if repo.Exist(name, password) {
		return entity.User{ID: 0, Name: name, Password: password}, nil
	}
	return entity.User{}, ErrNotFound
}

func (repo UserRepository) Set(name string, password string) (uint64, error) {
	return 0, nil
}

func (repo UserRepository) Exist(name string, password string) bool {
	return name == DummyUser.Name && password == DummyUser.Password
}

func (repo UserRepository) Find(id uint64) (entity.User, error) {
	return DummyUser, nil
}

var (
	DummyUser2 = entity.User{ID: 2, Name: "user2", Password: "password"}
	DummyUser3 = entity.User{ID: 3, Name: "user3", Password: "password"}

	DummyUserRelation = entity.UserRelation{
		Friends: []entity.User{DummyUser2, DummyUser3},
		Rooms:   DummyRooms,
	}
)

func (repo UserRepository) Relation(ctx context.Context, userID uint64) (entity.UserRelation, error) {
	return DummyUserRelation, nil
}
