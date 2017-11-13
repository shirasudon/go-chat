package inmemory

import (
	"github.com/shirasudon/go-chat/domain"
)

func OpenRepositories() *Repositories {
	return &Repositories{
		UserRepository:    &UserRepository{},
		MessageRepository: NewMessageRepository(),
		RoomRepository:    NewRoomRepository(),
		EventRepository:   &EventRepository{},
	}
}

type Repositories struct {
	*UserRepository
	*MessageRepository
	*RoomRepository
	*EventRepository
}

func (r Repositories) Users() domain.UserRepository {
	return r.UserRepository
}

func (r Repositories) Messages() domain.MessageRepository {
	return r.MessageRepository
}

func (r Repositories) Rooms() domain.RoomRepository {
	return r.RoomRepository
}

func (r Repositories) Events() domain.EventRepository {
	return r.EventRepository
}

func (r *Repositories) Close() error {
	return nil
}
