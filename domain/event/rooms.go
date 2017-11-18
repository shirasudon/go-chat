package event

// -----------------------
// Room events
// -----------------------

// Event for Room is created.
type RoomCreated struct {
	EventEmbd
	CreatedBy  uint64 `json:"created_by"`
	Name       string `json:"name"`
	IsTalkRoom bool
	MemberIDs  []uint64 `json:"member_ids"`
}

func (RoomCreated) Type() Type { return TypeRoomCreated }

// Event for Room is deleted.
type RoomDeleted struct {
	EventEmbd
	DeletedBy  uint64 `json:"deleted_by"`
	RoomID     uint64 `json:"room_id"`
	Name       string `json:"name"`
	IsTalkRoom bool
	MemberIDs  []uint64 `json:"member_ids"`
}

func (RoomDeleted) Type() Type { return TypeRoomDeleted }

// Event for Room added new member.
type RoomAddedMember struct {
	EventEmbd
	RoomID      uint64 `json:"room_id"`
	AddedUserID uint64 `json:"added_user_id"`
}

func (RoomAddedMember) Type() Type { return TypeRoomAddedMember }
