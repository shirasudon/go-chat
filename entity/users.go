package entity

import "context"

type User struct {
	ID       uint64
	Name     string
	Password string
}

type UserRelation struct {
	Friends []User
	Rooms   []Room
}

type UserRepository interface {
	FindByNameAndPassword(name string, password string) (User, error)
	ExistByNameAndPassword(name string, password string) bool

	Save(User) (uint64, error)

	Find(id uint64) (User, error)

	FindAllRelatedUsers(user_id uint64) ([]User, error)

	Relation(ctx context.Context, userID uint64) (UserRelation, error)
}
