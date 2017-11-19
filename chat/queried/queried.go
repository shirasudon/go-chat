// package queried contains queried results as
// Data-Transfer-Object.

package queried

import "time"

// Room is a detailed room information.
type Room struct {
	RoomName string `json:"room_name"`
	OwnerID  uint64 `json:"owner_id"`
	Members  []struct {
		UserID   uint64 `json:"user_id"`
		UserName string `json:"user_name"`
	} `json:"members"`
}

// UserRelation is the abstarct information associated with specified User.
type UserRelation struct {
	UserID   uint64 `json:"user_id"`
	UserName string `json:"user_name"`
	// TODO first name, last name

	Friends []UserFriend `json:"friends"`

	Rooms []UserRoom `json:"rooms"`
}

type UserFriend struct {
	UserID   uint64 `json:"user_id"`
	UserName string `json:"user_name"`
}

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
