package entity

import "context"

type User struct {
	ID       uint64
	Name     string
	Password string
}

type UserRepository interface {
	FindByNameAndPassword(ctx context.Context, name, password string) (User, error)
	ExistByNameAndPassword(ctx context.Context, name, password string) bool

	Store(context.Context, User) (uint64, error)

	Find(ctx context.Context, id uint64) (User, error)
}

type UserRelation struct {
	User
	Friends []User
	Rooms   []Room
}

type UserRelationRepository interface {
	Find(ctx context.Context, userID uint64) (UserRelation, error)
}
