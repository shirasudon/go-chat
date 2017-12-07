// package queried contains queried results as
// Data-Transfer-Object.

package queried

import "time"

// RoomInfo is a detailed room information.
// creator
type RoomInfo struct {
	RoomName    string        `json:"room_name"`
	RoomID      uint64        `json:"room_id"`
	CreatorID   uint64        `json:"room_creator_id"`
	Members     []UserProfile `json:"room_members"`
	MembersSize int           `json:"room_members_size"`
}

// UserRelation is the abstarct information associated with specified User.
type UserRelation struct {
	UserProfile

	Friends []UserProfile `json:"friends"`
	Rooms   []UserRoom    `json:"rooms"`
}

// UserProfile holds information for user profile.
type UserProfile struct {
	UserID    uint64 `json:"user_id"`
	UserName  string `json:"user_name"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

// UserRoom holds abstract information for the room.
type UserRoom struct {
	RoomID   uint64 `json:"room_id"`
	RoomName string `json:"room_name"`
}

// RoomMessages is a message list in specified Room.
type RoomMessages struct {
	RoomID uint64 `json:"room_id"`

	Msgs []Message `json:"messages"`

	Cursor struct {
		Current time.Time `json:"current"`
		Next    time.Time `json:"next"`
	} `json:"cursor"`
}

type Message struct {
	MessageID uint64    `json:"message_id"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

// UnreadRoomMessages is a list of unread messages in specified Room.
type UnreadRoomMessages struct {
	RoomID uint64 `json:"room_id"`

	Msgs     []Message `json:"messages"`
	MsgsSize int       `json:"messages_size"`
}
