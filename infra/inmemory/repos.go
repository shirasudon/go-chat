package inmemory

import (
	"github.com/shirasudon/go-chat/chat"
	"github.com/shirasudon/go-chat/domain"
	"github.com/shirasudon/go-chat/domain/event"
)

func OpenRepositories(pubsub chat.Pubsub) *Repositories {
	return &Repositories{
		UserRepository:    &UserRepository{},
		MessageRepository: NewMessageRepository(pubsub),
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

func (r Repositories) Events() event.EventRepository {
	return r.EventRepository
}

func (r *Repositories) Close() error {
	return nil
}
