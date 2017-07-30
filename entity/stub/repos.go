package stub

import (
	"github.com/mzki/chat/entity"
)

func init() {
	entity.RepositoryProducer = func(string) (entity.Repositories, error) {
		return repositories{}, nil
	}
}

type repositories struct{}

func (repositories) Users() entity.UserRepository {
	return UserRepository{}
}

func (repositories) Messages() entity.MessageRepository {
	panic("TODO: not implemented")
	return nil
}

func (repositories) Close() error { return nil }
