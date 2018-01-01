// package result provides result of the command service.

package result

// AddRoomMember is result for the chat.CommandService.AddRoomMember().
type AddRoomMember struct {
	RoomID uint64
	UserID uint64
}

// RemoveRoomMember is result for the chat.CommandService.RemoveRoomMember().
type RemoveRoomMember AddRoomMember
