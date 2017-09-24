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
	FindByNameAndPassword(ctx context.Context, name, password string) (User, error)
	ExistByNameAndPassword(ctx context.Context, name, password string) bool

	Save(context.Context, User) (uint64, error)

	Find(ctx context.Context, id uint64) (User, error)

	FindAllRelatedUsers(ctx context.Context, user_id uint64) ([]User, error)

	Relation(ctx context.Context, userID uint64) (UserRelation, error)
}
