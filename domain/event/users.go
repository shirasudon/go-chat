package event

// -----------------------
// User events
// -----------------------

// UserEventEmbd is EventEmbd with user event specific meta-data.
type UserEventEmbd struct {
	EventEmbd
}

func (UserEventEmbd) StreamID() StreamID { return UserStream }

// Event for User is created.
type UserCreated struct {
	UserEventEmbd
	Name      string   `json:"user_name"`
	FirstName string   `json:"first_name"`
	LastName  string   `json:"last_name"`
	FriendIDs []uint64 `json:"friend_ids"`
}

func (UserCreated) Type() Type { return TypeUserCreated }

// Event for User is created.
type UserAddedFriend struct {
	UserEventEmbd
	UserID        uint64 `json:"user_id"`
	AddedFriendID uint64 `json:"added_friend_id"`
}

func (UserAddedFriend) Type() Type { return TypeUserAddedFriend }
