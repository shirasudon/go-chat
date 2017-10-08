package domain

// Repositories holds any XXXRepository.
// user can get each repository from this.
type Repositories interface {
	Users() UserRepository
	Messages() MessageRepository
	Rooms() RoomRepository
}
