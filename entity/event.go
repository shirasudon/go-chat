package entity

type Event interface {
	EventType() EventType
}

type EventType uint

const (
	EventNone EventType = iota

	EventUserCreated = 10
	EventUserDeleted

	EventRoomCreated = 20
	EventRoomDeleted
	EventRoomAddedMember
	EventRoomRemoveMember
	EventRoomPostedMessage
	EventRoomUpdatedMessage
	EventRoomDeletedMessage
)

// Event for Room added new member.
type RoomAddedMember struct {
	RoomID      uint64
	AddedUserID uint64
}

func (RoomAddedMember) EventType() EventType { return EventRoomAddedMember }

// Event for Room posted new message.
type RoomPostedMessage struct {
	PostedRoomID uint64
	PostUserID   uint64
	Content      string
}

func (RoomPostedMessage) EventType() EventType { return EventRoomPostedMessage }
