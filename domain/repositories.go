//go:generate mockgen -destination=../mocks/mock_repos.go -package=mocks github.com/shirasudon/go-chat/domain Repositories

package domain

// Repositories holds any XXXRepository.
// you can get each repository from this.
type Repositories interface {
	Users() UserRepository
	Messages() MessageRepository
	Rooms() RoomRepository
}

// SimpleRepositories implementes Repositories interface.
// It acts just returning its fields when interface
// methods, Users(), Messages() and Rooms(), are called.
type SimpleRepositories struct {
	UserRepository    UserRepository
	MessageRepository MessageRepository
	RoomRepository    RoomRepository
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
