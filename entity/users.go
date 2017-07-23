package entity

type User struct {
	ID       int64
	UserID   string
	Password string
}

type UserRepository interface {
	Get(userID string, password string) (User, error)
	Set(userID string, password string) error
	Exist(userID string, password string) bool

	Find(id int64) (User, error)
}
