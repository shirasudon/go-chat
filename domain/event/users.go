package event

// -----------------------
// User events
// -----------------------

// Event for User is created.
type UserCreated struct {
	EventEmbd
	Name      string   `json:"user_name"`
	FriendIDs []uint64 `json:"friend_ids"`
}

func (UserCreated) Type() Type { return TypeUserCreated }

// Event for User is created.
type UserAddedFriend struct {
	EventEmbd
	UserID        uint64 `json:"user_id"`
	AddedFriendID uint64 `json:"added_friend_id"`
}

func (UserAddedFriend) Type() Type { return TypeUserAddedFriend }
