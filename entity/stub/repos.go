package stub

import (
	"github.com/mzki/go-chat/entity"
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
	return UserRepository{}
}

func (r repositories) Messages() entity.MessageRepository {
	return r.MessageRepository
}

func (repositories) Close() error { return nil }
