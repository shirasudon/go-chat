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

	userIDRelationMap = map[uint64]uint64{
		2: 3,
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

func (repo UserRepository) Save(ctx context.Context, u entity.User) (uint64, error) {
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

func (repo UserRepository) FindAllRelatedUsers(ctx context.Context, user_id uint64) ([]entity.User, error) {
	panic("TODO")
}

func (repo UserRepository) Relation(ctx context.Context, userID uint64) (entity.UserRelation, error) {
	r, ok := userRelationMap[userID]
	if ok {
		return r, nil
	}
	return DummyUser2Relation, ErrNotFound
}
