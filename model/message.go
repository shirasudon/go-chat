package model

import (
	"time"
)

// ActionMessage can return its action.
type ActionMessage interface {
	Action() Action
}

// common fields for the websocket action message structs.
// it implements ActionMessage interface.
type EmbdFields struct {
	ActionName Action `json:"action,omitempty"`
}

func (ef EmbdFields) Action() Action { return ef.ActionName }

// ToRoomMessage can return roomID for destination.
type ToRoomMessage interface {
	ToRoom() uint64
}

// common fields for the websocket message having
// destination as room.
// it implements ToRoomMessage interface.
type ToRoomFields struct {
	RoomID uint64 `json:"room_id,omitempty"`
}

func (tr ToRoomFields) ToRoom() uint64 {
	return tr.RoomID
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

	// front-end client to server
	ActionEnterRoom Action = "ENTER_ROOM"
	ActionExitRoom  Action = "EXIT_ROOM"

	// server from/to front-end client
	ActionReadMessage Action = "READ_MESSAGE"
	ActionChatMessage Action = "CHAT_MESSAGE"

	ActionTypeStart Action = "TYPE_START"
	ActionTypeEnd   Action = "TYPE_END"
)

// Error message.
// it implements ActionMessage interface.
type ErrorMessage struct {
	EmbdFields
	ErrorMsg string        `json:"error,omitempty"`
	Cause    ActionMessage `json:"cause,omitempty"`
}

func NewErrorMessage(err error, cause ...ActionMessage) ErrorMessage {
	em := ErrorMessage{EmbdFields: EmbdFields{ActionName: ActionError}, ErrorMsg: err.Error()}
	if len(cause) > 0 {
		em.Cause = cause[0]
	}
	return em
}

type UserConnect struct {
	EmbdFields
	UserID int `json:"user_id,omitempty"`
}

// EnterRoom indicates that user requests to enter
// specified room.
// it implements ActionMessage interface.
type EnterRoom struct {
	EmbdFields
	RoomID        uint64 `json:"enter_room_id,omitempty"`
	CurrentRoomID uint64 `json:"current_room_id,omitempty"`
	SenderID      uint64 `json:"sender_id,omitempty"`
}

func ParseEnterRoom(m AnyMessage, action Action) EnterRoom {
	if action != ActionEnterRoom {
		panic("ParseUserJoinRoom: invalid action")
	}
	v := EnterRoom{}
	v.ActionName = action
	v.SenderID, _ = m["sender_id"].(uint64)
	v.RoomID, _ = m["room_id"].(uint64)
	return v
}

// == ChatMessage related ActionMessages ==

// ChatMessage is chat message which is recieved from a browser-side
// client and sends to other clients in the same room.
// it implements ActionMessage and ToRoomMessage interface.
type ChatMessage struct {
	EmbdFields
	ToRoomFields
	ID       uint64 `json:"id,omitempty"` // used only server->client
	Content  string `json:"content,omitempty"`
	SenderID uint64 `json:"sender_id,omitempty"`
}

func ParseChatMessage(m AnyMessage, action Action) ChatMessage {
	if action != ActionChatMessage {
		panic("ParseChatMessage: invalid action")
	}
	cm := ChatMessage{}
	cm.ActionName = action
	cm.Content, _ = m["content"].(string)
	cm.SenderID, _ = m["sender_id"].(uint64)
	cm.RoomID, _ = m["room_id"].(uint64)
	return cm
}

// ReadMessage indicates notification which some chat messages are read by
// any user.
// it implements ActionMessage and ToRoomMessage interface.
type ReadMessage struct {
	EmbdFields
	ToRoomFields
	SenderID   uint64   `json:"sender_id,omitempty"`
	MessageIDs []uint64 `json:"message_ids"`
}

func ParseReadMessage(m AnyMessage, action Action) ReadMessage {
	if action != ActionReadMessage {
		panic("ParseReadMessage: invalid action")
	}
	rm := ReadMessage{}
	rm.ActionName = action
	rm.SenderID, _ = m["sender_id"].(uint64)
	rm.RoomID, _ = m["room_id"].(uint64)
	rm.MessageIDs, _ = m["message_ids"].([]uint64)
	return rm
}

// TypeStart indicates user starts key typing.
// it implements ActionMessage, ToRoomAction interface.
type TypeStart struct {
	EmbdFields
	ToRoomFields
	SenderID uint64 `json:"sender_id,omitempty"`

	// set by server and return client
	SenderName string    `json:"sender_name,omitempty"`
	StartAt    time.Time `json:"start_at,omitempty"`
}

func ParseTypeStart(m AnyMessage, action Action) TypeStart {
	if action != ActionTypeStart {
		panic("ParseTypeStart: invalid action")
	}
	ts := TypeStart{}
	ts.ActionName = action
	ts.SenderID, _ = m["sender_id"].(uint64)
	ts.RoomID, _ = m["room_id"].(uint64)
	return ts
}

// TypeEnd indicates user ends key typing.
// it implements ActionMessage, ToRoomMessage interface.
type TypeEnd struct {
	EmbdFields
	ToRoomFields
	SenderID uint64 `json:"sender_id,omitempty"`

	// set by server and return client
	SenderName string    `json:"sender_name,omitempty"`
	EndAt      time.Time `json:"end_at,omitempty"`
}

func ParseTypeEnd(m AnyMessage, action Action) TypeEnd {
	if action != ActionTypeEnd {
		panic("ParseTypeEnd: invalid action")
	}
	te := TypeEnd{}
	te.ActionName = action
	te.SenderID, _ = m["sender_id"].(uint64)
	te.RoomID, _ = m["room_id"].(uint64)
	return te
}
