package sqlite

import (
	"database/sql"

	"github.com/jmoiron/sqlx"
	"github.com/mzki/chat/entity"
)

// UserRepository manages access to user table in sqlite database.
// It must be singleton object since database conncetion is so.
type UserRepository struct {
	db *sqlx.DB
}

var (
	findByNameAndPassword   *sqlx.Stmt
	insertByNameAndPassword *sqlx.Stmt
)

const (
	queryFindByNameAndPassword = `SELECT id, name, password FROM users
	WHERE name =$1 and password =$2`
	queryInsertByNameAndPassword = `INSERT INTO users (name, password) VALUES ($1, $2)`
)

func newUserRepository(db *sqlx.DB) (*UserRepository, error) {
	var err error
	findByNameAndPassword, err = db.Preparex(queryFindByNameAndPassword)
	if err != nil {
		return err
	}
	insertByNameAndPassword, err = db.Preparex(queryInsertByNameAndPassword)
	if err != nil {
		return err
	}
	return &UserRepository{
		db: db,
	}
}

func (repo *UserRepository) Get(name string, password string) (entity.User, error) {
	panic("not implemented")
}

func (repo *UserRepository) Set(name string, password string) error {
	panic("not implemented")
}

func (repo *UserRepository) Exist(name string, password string) bool {
	panic("not implemented")
}

func (repo *UserRepository) Find(id int64) (entity.User, error) {
	panic("not implemented")
}
