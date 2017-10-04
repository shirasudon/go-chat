package entity

import "context"

type User struct {
	ID       uint64
	Name     string
	Password string
}

type UserRepository interface {
	TxBeginner

	FindByNameAndPassword(ctx context.Context, name, password string) (User, error)
	ExistByNameAndPassword(ctx context.Context, name, password string) bool

	// Store specified user to the repository, and return user id
	// for stored new user.
	Store(context.Context, User) (uint64, error)

	// Find one user by id.
	Find(ctx context.Context, id uint64) (User, error)

	// Find all users related with the specified user id.
	FindAllByUserID(ctx context.Context, userID uint64) ([]User, error)
}
