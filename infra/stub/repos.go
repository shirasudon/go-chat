package stub

import (
	"github.com/shirasudon/go-chat/domain"
)

func init() {
	domain.RepositoryProducer = func(string) (domain.Repositories, error) {
		return repositories{
			MessageRepository: newMessageRepository(),
		}, nil
	}
}

type repositories struct {
	domain.MessageRepository
}

func (repositories) Users() domain.UserRepository {
	return &UserRepository{}
}

func (r repositories) Messages() domain.MessageRepository {
	return r.MessageRepository
}

func (r repositories) Rooms() domain.RoomRepository {
	return &RoomRepository{}
}
