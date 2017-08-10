package stub

import (
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
