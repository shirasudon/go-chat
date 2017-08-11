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
	Get(name string, password string) (User, error)
	Set(name string, password string) (uint64, error)
	Exist(name string, password string) bool

	Find(id uint64) (User, error)

	Relation(ctx context.Context, userID uint64) (UserRelation, error)
}
