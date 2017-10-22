package domain

import (
	"context"
	"fmt"
)

//go:generate mockgen -destination=../mocks/mock_users.go -package=mocks github.com/shirasudon/go-chat/domain UserRepository

type UserRepository interface {
	TxBeginner

	FindByNameAndPassword(ctx context.Context, name, password string) (User, error)
	ExistByNameAndPassword(ctx context.Context, name, password string) bool

	// Store specified user to the repository, and return user id
	// for stored new user.
	Store(context.Context, User) (uint64, error)

	// Find one user by id.
	Find(ctx context.Context, id uint64) (User, error)

	// Find all users related with the specified user id.
	FindAllByUserID(ctx context.Context, userID uint64) ([]User, error)
}

// set for user id.
type UserIDSet struct {
	idMap map[uint64]bool
}

func NewUserIDSet(ids ...uint64) UserIDSet {
	idMap := make(map[uint64]bool, len(ids))
	for _, id := range ids {
		idMap[id] = true
	}
	return UserIDSet{idMap}
}

func (set *UserIDSet) getIDMap() map[uint64]bool {
	if set.idMap == nil {
		set.idMap = make(map[uint64]bool, 4)
	}
	return set.idMap
}

func (set *UserIDSet) Has(id uint64) bool {
	_, ok := set.getIDMap()[id]
	return ok
}

func (set *UserIDSet) Add(id uint64) {
	set.getIDMap()[id] = true
}

func (set *UserIDSet) Remove(id uint64) {
	delete(set.getIDMap(), id)
}

// It returns a deep copy of the ID list.
func (set *UserIDSet) List() []uint64 {
	idMap := set.getIDMap()
	ids := make([]uint64, 0, len(idMap))
	for id, _ := range idMap {
		ids = append(ids, id)
	}
	return ids
}

// User entity. Its fields are exported
// due to construct from the datastore.
// In application side, creating/modifying/deleting the user
// should be done by the methods which emits the domain event.
type User struct {
	EventHolder

	ID       uint64
	Name     string
	Password string

	FriendIDs UserIDSet
}

// create new Room entity into the repository. It retruns the new user
// holding event for UserCreated and error if any.
func NewUser(ctx context.Context, userRepo UserRepository, name string, password string, friendIDs UserIDSet) (User, error) {
	u := User{
		EventHolder: NewEventHolder(),
		ID:          0, // 0 means new entity
		Name:        name,
		Password:    password,
		FriendIDs:   friendIDs,
	}

	id, err := userRepo.Store(ctx, u)
	if err != nil {
		return u, err
	}
	u.ID = id

	ev := UserCreated{
		Name:      name,
		Password:  password,
		FriendIDs: friendIDs.List(),
	}
	u.AddEvent(ev)

	return u, nil
}

func (u *User) IsNew() bool { return u.ID == 0 }

// It adds the friend to the user.
// It returns the event adding into the user, and error
// when the friend already exist in the user.
func (u *User) AddFriend(friend User) (UserAddedFriend, error) {
	if u.ID == 0 {
		return UserAddedFriend{}, fmt.Errorf("newly user can not be added friend")
	}
	if u.ID == friend.ID {
		return UserAddedFriend{}, fmt.Errorf("can not add user itself as friend")
	}
	if u.HasFriend(friend) {
		return UserAddedFriend{}, fmt.Errorf("friend(id=%d) already exist in the user(id=%d)", friend.ID, u.ID)
	}

	u.FriendIDs.Add(friend.ID)

	ev := UserAddedFriend{
		UserID:        u.ID,
		AddedFriendID: friend.ID,
	}
	u.AddEvent(ev)
	return ev, nil
}

// It returns whether the user has specified friend?
func (u *User) HasFriend(friend User) bool {
	return u.FriendIDs.Has(friend.ID)
}

// -----------------------
// User events
// -----------------------

// Event for User is created.
type UserCreated struct {
	Name      string
	Password  string
	FriendIDs []uint64
}

func (UserCreated) EventType() EventType { return EventUserCreated }

// Event for User is created.
type UserAddedFriend struct {
	UserID        uint64
	AddedFriendID uint64
}

func (UserAddedFriend) EventType() EventType { return EventUserAddedFriend }
