package entity

type User struct {
	ID       uint
	Name     string
	Password string
}

type UserRepository interface {
	Get(name string, password string) (User, error)
	Set(name string, password string) error
	Exist(name string, password string) bool

	Find(id uint) (User, error)
}
