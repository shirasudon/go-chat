package sqlite3

import (
	"context"
	"errors"

	"github.com/jmoiron/sqlx"
	"github.com/shirasudon/go-chat/entity"
)

// UserRepository manages access to user table in sqlite database.
// It must be singleton object since database conncetion is so.
type UserRepository struct {
	db                    *sqlx.DB
	findByID              *sqlx.Stmt
	findByNameAndPassword *sqlx.Stmt
	insertUser            *sqlx.Stmt

	findFriendsByUserID *sqlx.Stmt

	rooms entity.RoomRepository
}

const (
	userQueryFindByID = `
SELECT * FROM users where id=$1 LIMIT 1`
	userQueryFindByNameAndPassword = `
SELECT * FROM users WHERE name=$1 and password=$2 LIMIT 1`
	userNamedQueryInsertByUser = `
	INSERT INTO users (name, password) VALUES (:name, :password)`

	userQueryFindFriendsByUserID = `
SELECT * FROM users INNER JOIN user_friends ON users.id = user_friends.user_id
 WHERE users.id = $1 ORDER BY users.id ASC`
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
	insertUser, err := db.Preparex(userNamedQueryInsertByUser)
	if err != nil {
		return nil, err
	}
	findFriendsByUserID, err := db.Preparex(userQueryFindFriendsByUserID)
	if err != nil {
		return nil, err
	}

	return &UserRepository{
		db:                    db,
		findByID:              findByID,
		findByNameAndPassword: findByNameAndPassword,
		insertUser:            insertUser,
		findFriendsByUserID:   findFriendsByUserID,
	}, nil
}

func (repo *UserRepository) close() {
	for _, stmt := range []*sqlx.Stmt{
		repo.findByID,
		repo.findByNameAndPassword,
		repo.insertUser,
		repo.findFriendsByUserID,
	} {
		stmt.Close()
	}
	repo.rooms = nil
}

func (repo *UserRepository) FindByNameAndPassword(name string, password string) (entity.User, error) {
	u := entity.User{}
	err := repo.findByNameAndPassword.Get(&u, name, password)
	return u, err
}

func (repo *UserRepository) Save(u entity.User) (uint64, error) {
	res, err := repo.insertUser.Exec(&u)
	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()
	return uint64(id), err
}

func (repo *UserRepository) ExistByNameAndPassword(name string, password string) bool {
	_, err := repo.FindByNameAndPassword(name, password)
	return err == nil
}

func (repo *UserRepository) Find(id uint64) (entity.User, error) {
	u := entity.User{}
	err := repo.findByID.Get(&u, id)
	return u, err
}

func (repo *UserRepository) Relation(ctx context.Context, userID uint64) (entity.UserRelation, error) {
	var relaion entity.UserRelation
	// validate existance of repo.rooms to use it.
	if repo.rooms == nil {
		return relaion, errors.New("UserRepository does not have RoomRepository, be sure set it")
	}

	if err := repo.findFriendsByUserID.SelectContext(ctx, &(relaion.Friends), userID); err != nil {
		return relaion, err
	}
	var err error
	relaion.Rooms, err = repo.rooms.GetUserRooms(ctx, userID)
	return relaion, err
}
