package stub

import (
	"errors"

	"github.com/mzki/chat/entity"
)

type UserRepository struct{}

var ErrNotFound = errors.New("not found")

func (repo UserRepository) Get(userID string, password string) (entity.User, error) {
	if repo.Exist(userID, password) {
		return entity.User{ID: 0, UserID: userID, Password: password}, nil
	}
	return entity.User{}, ErrNotFound
}

func (repo UserRepository) Set(userID string, password string) error {
	return nil
}

func (repo UserRepository) Exist(userID string, password string) bool {
	return userID == "user" && password == "password"
}

func (repo UserRepository) Find(id int64) (entity.User, error) {
	return entity.User{}, ErrNotFound
}
