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

// ChatActionMessage is used for chat context, which has
// roomID and senderID(userID) for destination.
// it also implements ActionMessage interface.
type ChatActionMessage interface {
	ActionMessage
	GetRoomID() uint64
	GetSenderID() uint64
}

// common fields for the websocket message to be
// used to chat context.
// it implements ChatActionMessage interface.
type ChatActionFields struct {
	EmbdFields
	RoomID   uint64 `json:"room_id,omitempty"`
	SenderID uint64 `json:"sender_id,omitempty"` // it is overwritten by the server
}

func (tr ChatActionFields) GetRoomID() uint64 {
	return tr.RoomID
}

func (tr ChatActionFields) GetSenderID() uint64 {
	return tr.SenderID
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

func (a AnyMessage) String(key string) string {
	n, _ := a[key].(string)
	return n
}

func (a AnyMessage) Number(key string) float64 {
	n, _ := a[key].(float64)
	return n
}

func (a AnyMessage) Array(key string) []interface{} {
	n, _ := a[key].([]interface{})
	return n
}

func (a AnyMessage) Object(key string) map[string]interface{} {
	n, _ := a[key].(map[string]interface{})
	return n
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

// UserConnect indicates connect acitve user to chat server.
// it implements ActionMessage interface
type UserConnect struct {
	EmbdFields
	UserID uint64 `json:"user_id,omitempty"`
}

func NewUserConnect(userID uint64) UserConnect {
	return UserConnect{
		EmbdFields: EmbdFields{
			ActionName: ActionUserConnect,
		},
		UserID: userID,
	}
}

// UserDisconnect indicates disconnect acitve user to chat server.
// it implements ActionMessage interface
type UserDisconnect UserConnect

func NewUserDisconnect(userID uint64) UserDisconnect {
	return UserDisconnect(NewUserConnect(userID))
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
	v.SenderID = uint64(m.Number("sender_id"))
	v.RoomID = uint64(m.Number("room_id"))
	return v
}

// == ChatMessage related ActionMessages ==

// ChatMessage is chat message which is recieved from a browser-side
// client and sends to other clients in the same room.
// it implements ChatActionMessage interface.
type ChatMessage struct {
	ChatActionFields
	ID      uint64 `json:"id,omitempty"` // used only server->client
	Content string `json:"content,omitempty"`
}

func ParseChatMessage(m AnyMessage, action Action) ChatMessage {
	if action != ActionChatMessage {
		panic("ParseChatMessage: invalid action")
	}
	cm := ChatMessage{}
	cm.ActionName = action
	cm.Content = m.String("content")
	cm.RoomID = uint64(m.Number("room_id"))
	return cm
}

// ReadMessage indicates notification which some chat messages are read by
// any user.
// it implements ChatActionMessage interface.
type ReadMessage struct {
	ChatActionFields
	MessageIDs []uint64 `json:"message_ids"`
}

func ParseReadMessage(m AnyMessage, action Action) ReadMessage {
	if action != ActionReadMessage {
		panic("ParseReadMessage: invalid action")
	}
	rm := ReadMessage{}
	rm.ActionName = action
	rm.RoomID = uint64(m.Number("room_id"))
	anys := m.Array("message_ids")
	msg_ids := make([]uint64, 0, len(anys))
	for _, v := range anys {
		if n, ok := v.(float64); ok {
			msg_ids = append(msg_ids, uint64(n))
		}
	}
	rm.MessageIDs = msg_ids
	return rm
}

// TypeStart indicates user starts key typing.
// it implements ChatActionMessage interface.
type TypeStart struct {
	ChatActionFields

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
	ts.RoomID = uint64(m.Number("room_id"))
	return ts
}

// TypeEnd indicates user ends key typing.
// it implements ChatActionMessage interface.
type TypeEnd struct {
	ChatActionFields

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
	te.RoomID = uint64(m.Number("room_id"))
	return te
}
