package chat

import (
	"time"
)

// AnyMessage is a arbitrary message through the websocket.
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

	// server to front-end client
	ActionUserConnect    Action = "USER_CONNECT"
	ActionUserDisconnect Action = "USER_DISCONNECT"
	ActionCreateRoom     Action = "CREATE_ROOM"

	// server from/to front-end client
	ActionReadMessage Action = "READ_MESSAGE"
	ActionChatMessage Action = "CHAT_MESSAGE"
	ActionTypeStart   Action = "TYPE_START"
	ActionTypeEnd     Action = "TYPE_END"
)

// common fields for the websocket payload structs.
type embdFields struct {
	Action `json:"action",omitempty`
}

// ChatMessage is chat message which is recieved from a browser-side
// client and sends to other clients in the same room.
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

type UserConnect struct {
	embdFields
	UserID int `json:"user_id,omitempty"`
}
