package entity

type Event interface {
	EventType() EventType
}

type EventType uint

const (
	EventNone EventType = iota

	EventUserCreated = 10
	EventUserDeleted
	EventUserAddedFriend

	EventRoomCreated = 20
	EventRoomDeleted
	EventRoomAddedMember
	EventRoomRemoveMember
	EventRoomPostedMessage
	EventRoomUpdatedMessage
	EventRoomDeletedMessage
)

// EventHolder holds event objects.
// It is used to embed into entity.
type EventHolder struct {
	events []Event
}

func NewEventHolder() EventHolder {
	return EventHolder{
		events: make([]Event, 0, 2),
	}
}

func (holder *EventHolder) Events() []Event {
	if holder.events == nil {
		holder.events = make([]Event, 0, 2)
	}
	newEvents := make([]Event, 0, len(holder.events))
	for _, ev := range holder.events {
		newEvents = append(newEvents, ev)
	}
	return newEvents
}

func (holder *EventHolder) AddEvent(ev Event) {
	if holder.events == nil {
		holder.events = make([]Event, 0, 2)
	}
	holder.events = append(holder.events, ev)
}

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

// Event for Room is created.
type RoomCreated struct {
	Name       string
	IsTalkRoom bool
	MemberIDs  []uint64
}

func (RoomCreated) EventType() EventType { return EventRoomCreated }

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
