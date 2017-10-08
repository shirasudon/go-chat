package sqlite3

import (
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"github.com/shirasudon/go-chat/entity"
)

func init() {
	entity.RepositoryProducer = RepositoryProducer
}

func RepositoryProducer(dataSourceName string) (entity.Repositories, error) {
	db, err := sqlx.Open("sqlite3", dataSourceName)
	if err != nil {
		return nil, err
	}

	uRepo, err := newUserRepository(db)
	if err != nil {
		return nil, err
	}

	rRepo, err := newRoomRepository(db)
	if err != nil {
		return nil, err
	}
	uRepo.rooms = rRepo
	rRepo.users = uRepo

	return &Repositories{
		DB:                db,
		UserRepository:    uRepo,
		MessageRepository: nil, // TODO implement
		RoomRepository:    rRepo,
	}, nil
}

type Repositories struct {
	DB *sqlx.DB
	*UserRepository
	*MessageRepository
	*RoomRepository
}

func (r Repositories) Users() entity.UserRepository {
	return r.UserRepository
}

func (r Repositories) Messages() entity.MessageRepository {
	panic("entity/sqlite: TODO: not implement")
	return r.MessageRepository
}

func (r Repositories) Rooms() entity.RoomRepository {
	panic("entity/sqlite: TODO: not implement")
	return r.RoomRepository
}

func (r Repositories) BeginTx(ctx context.Context) (entity.Tx, error) {
	return r.DB.BeginTxx(ctx, nil)
}

func (r Repositories) Close() error {
	r.UserRepository.close()
	return r.DB.Close()
}
