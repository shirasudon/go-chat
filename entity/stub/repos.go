package stub

import (
	"context"
	"database/sql"

	"github.com/shirasudon/go-chat/entity"
)

func init() {
	entity.RepositoryProducer = func(string) (entity.Repositories, error) {
		return repositories{
			MessageRepository: newMessageRepository(),
		}, nil
	}
}

type repositories struct {
	entity.MessageRepository
}

func (repositories) Users() entity.UserRepository {
	return &UserRepository{}
}

func (r repositories) Messages() entity.MessageRepository {
	return r.MessageRepository
}

func (r repositories) Rooms() entity.RoomRepository {
	return &RoomRepository{}
}

func (r repositories) RoomRelations() entity.RoomRelationRepository {
	return &RoomRelationRepository{}
}

func (r repositories) BeginTx(ctx context.Context, opt *sql.TxOptions) (entity.Tx, error) {
	return txStub{}, nil
}

type txStub struct{}

func (txStub) Commit() error   { return nil }
func (txStub) Rollback() error { return nil }

func (repositories) Close() error { return nil }
