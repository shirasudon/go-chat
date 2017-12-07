package action

import "time"

// QueryRoomMessages is a query for
// messages in specified room.
type QueryRoomMessages struct {
	RoomID uint64    `json:"room_id"`
	Before time.Time `json:"before"`
	Limit  int       `json:"limit"`
}

// QueryUnreadRoomMessages is a query for
// unread messages by user in specified room.
type QueryUnreadRoomMessages struct {
	RoomID uint64 `json:"room_id"`
	Limit  int    `json:"limit"`
}
