package model

import (
	"time"
)

// ActionMessage can return its action.
type ActionMessage interface {
	Action() Action
}

// AnyMessage is a arbitrary message through the websocket.
// it implements ActionMessage interface.
type AnyMessage map[string]interface{}

// key for the action field in AnyMessage.
const KeyAction = "action"

// get action from any message which indicates
// what action is contained any message.
// return empty action if no action exist.
func (a AnyMessage) Action() Action {
	if action, ok := a[KeyAction].(string); ok {
		return Action(action)
	}
	return ActionEmpty
}

// Action indicates a action type for the JSON data schema.
type Action string

const (
	// no meaning action
	ActionEmpty Action = ""

	// internal server error
	ActionError Action = "ERROR"

	// server to front-end client
	ActionUserConnect    Action = "USER_CONNECT"
	ActionUserDisconnect Action = "USER_DISCONNECT"

	ActionCreateRoom Action = "CREATE_ROOM"
	ActionDeleteRoom Action = "DELETE_ROOM"

	// server from/to front-end client
	ActionReadMessage Action = "READ_MESSAGE"
	ActionChatMessage Action = "CHAT_MESSAGE"

	ActionTypeStart Action = "TYPE_START"
	ActionTypeEnd   Action = "TYPE_END"

	ActionUserJoinRoom  Action = "USER_JOIN_ROOM"
	ActionUserLeaveRoom Action = "USER_LEAVE_ROOM"
)

// common fields for the websocket payload structs.
type embdFields struct {
	Action `json:"action,omitempty"`
}

func (ef embdFields) Action() Action { return ef.Action }

// Error message.
// it implements ActionMessage interface.
type ErrorMessage struct {
	embdFields
	Error error `json:"error"`
}

func NewErrorMessage(err error) ErrorMessage {
	return ErrorMessage{Action: ActionError, Error: err}
}

// ChatMessage is chat message which is recieved from a browser-side
// client and sends to other clients in the same room.
// it implements ActionMessage interface.
type ChatMessage struct {
	embdFields
	ID       uint64 `json:"id,omitempty"` // used only server->client
	Content  string `json:"content,omitempty"`
	SenderID uint64 `json:"sender_id,omitempty"`
	RoomID   uint64 `json:"room_id,omitempty"`
}

func ParseChatMessage(m AnyMessage, action Action) ChatMessage {
	if action != ActionChatMessage {
		panic("ParseChatMessage: invalid action")
	}
	cm := ChatMessage{}
	cm.Action = action
	cm.Content, _ = m["content"].(string)
	cm.SenderID, _ = m["sender_id"].(uint64)
	cm.RoomID, _ = m["room_id"].(uint64)
	return cm
}

// ReadMessage indicates notification which some chat messages are read by
// any user.
// it implements ActionMessage interface.
type ReadMessage struct {
	embdFields
	SenderID   uint64   `json:"sender_id,omitempty"`
	MessageIDs []uint64 `json:"message_ids,omitempty"`
}

func ParseReadMessage(m AnyMessage, action Action) ReadMessage {
	if action != ActionReadMessage {
		panic("ParseReadMessage: invalid action")
	}
	rm := ReadMessage{}
	rm.Action = action
	rm.SenderID, _ = m["sender_id"].(uint64)
	rm.MessageIDs, _ = m["message_ids"].([]uint64)
	return rm
}

// TypeStart indicates user starts key typing.
// it implements ActionMessage interface.
type TypeStart struct {
	embdFields
	SenderID   uint64    `json:"sender_id,omitempty"`
	SenderName string    `json:"sender_name,omitempty"`
	StartAt    time.Time `json:"start_at,omitempty"`
}

func ParseTypeStart(m AnyMessage, action Action) TypeStart {
	if action != ActionTypeStart {
		panic("ParseTypeStart: invalid action")
	}
	ts := TypeStart{}
	ts.Action = action
	// ts.SenderID, _ = m["sender_id"].(uint)
	// ts.SenderName, _ = m["sender_name"].(string)
	// ts.StartAt, _ = m["start_at"].(time.Time)
	return ts
}

// TypeEnd indicates user ends key typing.
// it implements ActionMessage interface.
type TypeEnd struct {
	embdFields
	SenderID   uint64    `json:"sender_id,omitempty"`
	SenderName string    `json:"sender_name,omitempty"`
	EndAt      time.Time `json:"end_at,omitempty"`
}

func ParseTypeEnd(m AnyMessage, action Action) TypeEnd {
	if action != ActionTypeEnd {
		panic("ParseTypeEnd: invalid action")
	}
	te := TypeEnd{}
	te.Action = action
	// te.SenderID, _ = m["sender_id"].(uint)
	// te.SenderName, _ = m["sender_name"].(string)
	// te.EndAt, _ = m["end_at"].(time.Time)
	return te
}

// UserJoinRoom indicates that user requests to join
// specified room.
// it implements ActionMessage interface.
type UserJoinRoom struct {
	embdFields
	SenderID uint64 `json:"sender_id,omitempty"`
	RoomID   uint64 `json:"room_id,omitempty"`
}

func ParseUserJoinRoom(m AnyMessage, action Action) UserJoinRoom {
	if action != ActionUserJoinRoom {
		panic("ParseUserJoinRoom: invalid action")
	}
	v := UserJoinRoom{}
	v.Action = action
	v.SenderID, _ = m["sender_id"].(uint64)
	v.RoomID, _ = m["room_id"].(uint64)
	return v
}

type UserConnect struct {
	embdFields
	UserID int `json:"user_id,omitempty"`
}
