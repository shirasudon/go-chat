package stub

import (
	"github.com/shirasudon/go-chat/domain"
)

func OpenRepositories() *Repositories {
	return &Repositories{
		MessageRepository: NewMessageRepository(),
		RoomRepository:    NewRoomRepository(),
	}
}

type Repositories struct {
	domain.MessageRepository
	domain.RoomRepository
}

func (Repositories) Users() domain.UserRepository {
	return &UserRepository{}
}

func (r Repositories) Messages() domain.MessageRepository {
	return r.MessageRepository
}

func (r Repositories) Rooms() domain.RoomRepository {
	return &RoomRepository{}
}

func (r *Repositories) Close() error {
	return nil
}
