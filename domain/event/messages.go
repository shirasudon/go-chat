package event

// -----------------------
// Message events
// -----------------------

// MessageEventEmbd is EventEmbd with message event specific meta-data.
type MessageEventEmbd struct {
	EventEmbd
}

func (MessageEventEmbd) StreamID() StreamID { return MessageStream }

// Event for the message is created.
type MessageCreated struct {
	MessageEventEmbd
	MessageID uint64 `json:"message_id"`
	RoomID    uint64 `json:"room_id"`
	CreatedBy uint64 `json:"created_by"`
	Content   string `json:"content"`
}

func (MessageCreated) Type() Type { return TypeMessageCreated }

// Event for the message is read by the user.
type MessageReadByUser struct {
	MessageEventEmbd
	MessageID uint64 `json:"message_id"`
	UserID    uint64 `json:"user_id"`
}

func (MessageReadByUser) Type() Type { return TypeMessageReadByUser }
