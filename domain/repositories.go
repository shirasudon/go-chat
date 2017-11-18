package domain

import "github.com/shirasudon/go-chat/domain/event"

//go:generate mockgen -destination=../internal/mocks/mock_repos.go -package=mocks github.com/shirasudon/go-chat/domain Repositories

// Repositories holds any XXXRepository.
// you can get each repository from this.
type Repositories interface {
	Users() UserRepository
	Messages() MessageRepository
	Rooms() RoomRepository

	Events() event.EventRepository
}

// SimpleRepositories implementes Repositories interface.
// It acts just returning its fields when interface
// methods, Users(), Messages() and Rooms(), are called.
type SimpleRepositories struct {
	UserRepository    UserRepository
	MessageRepository MessageRepository
	RoomRepository    RoomRepository

	EventRepository event.EventRepository
}

func (s SimpleRepositories) Users() UserRepository {
	return s.UserRepository
}

func (s SimpleRepositories) Messages() MessageRepository {
	return s.MessageRepository
}

func (s SimpleRepositories) Rooms() RoomRepository {
	return s.RoomRepository
}

func (s SimpleRepositories) Events() event.EventRepository {
	return s.EventRepository
}
