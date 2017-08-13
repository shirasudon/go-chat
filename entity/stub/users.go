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

	DummyUser2Relation = entity.UserRelation{
		Friends: []entity.User{DummyUser3},
		Rooms:   []entity.Room{DummyRoom2, DummyRoom3},
	}

	userMap = map[uint64]entity.User{
		0: DummyUser,
		2: DummyUser2,
		3: DummyUser3,
	}

	userRelationMap = map[uint64]entity.UserRelation{
		2: DummyUser2Relation,
	}
)

func (repo UserRepository) Get(name string, password string) (entity.User, error) {
	if repo.Exist(name, password) {
		return entity.User{ID: 0, Name: name, Password: password}, nil
	}
	return entity.User{}, ErrNotFound
}

func (repo UserRepository) Set(name string, password string) (uint64, error) {
	panic("not implement")
	return 0, nil
}

func (repo UserRepository) Exist(name string, password string) bool {
	for _, u := range userMap {
		if name == u.Name && password == u.Password {
			return true
		}
	}
	return false
}

func (repo UserRepository) Find(id uint64) (entity.User, error) {
	u, ok := userMap[id]
	if ok {
		return u, nil
	}
	return DummyUser, ErrNotFound
}

func (repo UserRepository) Relation(ctx context.Context, userID uint64) (entity.UserRelation, error) {
	r, ok := userRelationMap[userID]
	if ok {
		return r, nil
	}
	return DummyUser2Relation, ErrNotFound
}
