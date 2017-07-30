package sqlite3

import (
	"github.com/jmoiron/sqlx"
	"github.com/mzki/chat/entity"
)

// UserRepository manages access to user table in sqlite database.
// It must be singleton object since database conncetion is so.
type UserRepository struct {
	db                      *sqlx.DB
	findByID                *sqlx.Stmt
	findByNameAndPassword   *sqlx.Stmt
	insertByNameAndPassword *sqlx.Stmt
}

const (
	userQueryFindByID              = `SELECT * FROM users where id=$1`
	userQueryFindByNameAndPassword = `SELECT * FROM users
																	WHERE name=$1 and password=$2`
	userQueryInsertByNameAndPassword = `INSERT INTO users (name, password) VALUES ($1, $2)`
)

func newUserRepository(db *sqlx.DB) (*UserRepository, error) {
	findByID, err := db.Preparex(userQueryFindByID)
	if err != nil {
		return nil, err
	}
	findByNameAndPassword, err := db.Preparex(userQueryFindByNameAndPassword)
	if err != nil {
		return nil, err
	}
	insertByNameAndPassword, err := db.Preparex(userQueryInsertByNameAndPassword)
	if err != nil {
		return nil, err
	}

	return &UserRepository{
		db:                      db,
		findByID:                findByID,
		findByNameAndPassword:   findByNameAndPassword,
		insertByNameAndPassword: insertByNameAndPassword,
	}, nil
}

func (repo *UserRepository) Get(name string, password string) (entity.User, error) {
	u := entity.User{}
	err := repo.findByNameAndPassword.Get(&u, name, password)
	return u, err
}

func (repo *UserRepository) Set(name string, password string) error {
	_, err := repo.insertByNameAndPassword.Exec(name, password)
	return err
}

func (repo *UserRepository) Exist(name string, password string) bool {
	_, err := repo.Get(name, password)
	return err == nil
}

func (repo *UserRepository) Find(id uint) (entity.User, error) {
	u := entity.User{}
	err := repo.findByID.Get(&u, id)
	return u, err
}

func (repo *UserRepository) close() {
	for _, stmt := range []*sqlx.Stmt{
		repo.findByID,
		repo.findByNameAndPassword,
		repo.insertByNameAndPassword,
	} {
		stmt.Close()
	}
}
