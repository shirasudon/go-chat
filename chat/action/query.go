package action

// QueryRoomMessages is a query for
// messages in specified room.
type QueryRoomMessages struct {
	RoomID uint64
	Before Timestamp `json:"before" query:"before"`
	Limit  int       `json:"limit" query:"limit"`
}

// QueryUnreadRoomMessages is a query for
// unread messages by user in specified room.
type QueryUnreadRoomMessages struct {
	RoomID uint64
	Limit  int `json:"limit" query:"limit"`
}
