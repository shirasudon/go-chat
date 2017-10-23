package action

import (
	"errors"
	"time"
)

// ActionMessage can return its action.
type ActionMessage interface {
	Action() Action
}

// AnyMessage is a arbitrary message through the websocket.
// it implements ActionMessage interface.
type AnyMessage map[string]interface{}

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

func (a AnyMessage) SetString(key string, val string) {
	a[key] = val
}

func (a AnyMessage) Number(key string) float64 {
	n, _ := a[key].(float64)
	return n
}

func (a AnyMessage) SetNumber(key string, val float64) {
	a[key] = val
}

func (a AnyMessage) Array(key string) []interface{} {
	n, _ := a[key].([]interface{})
	return n
}

func (a AnyMessage) UInt64s(key string) []uint64 {
	anys := a.Array(key)
	uint64s := make([]uint64, 0, len(anys))
	for _, v := range anys {
		if n, ok := v.(float64); ok {
			uint64s = append(uint64s, uint64(n))
		}
	}
	return uint64s
}

func (a AnyMessage) Object(key string) map[string]interface{} {
	n, _ := a[key].(map[string]interface{})
	return n
}

// Convert AnyMessage to ActionMessage specified by
// AnyMessage.Action().
// it returns error if AnyMessage has invalid data structure.
func ConvertAnyMessage(m AnyMessage) (ActionMessage, error) {
	a := m.Action()
	switch a {
	case ActionChatMessage:
		return ParseChatMessage(m, a)
	case ActionReadMessage:
		return ParseReadMessage(m, a)
	case ActionTypeStart:
		return ParseTypeStart(m, a)
	case ActionTypeEnd:
		return ParseTypeEnd(m, a)
	case ActionEmpty:
		return m, errors.New("JSON object must have any action field")
	}
	return m, errors.New("unknown action: " + string(a))
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
)

const (
	// key for the action field in AnyMessage.
	KeyAction   = "action"
	KeySenderID = "sender_id"
	KeyRoomID   = "room_id"
)

// common fields for the websocket action message structs.
// it implements ActionMessage interface.
type EmbdFields struct {
	ActionName Action `json:"action,omitempty"`
}

func (ef EmbdFields) Action() Action { return ef.ActionName }

// helper function for parsing fields from AnyMessage.
// it will load Action from AnyMessage.
func (ef *EmbdFields) ParseFields(m AnyMessage) {
	ef.ActionName = m.Action()
}

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

// helper function for parsing fields from AnyMessage.
// it will load RoomID and SenderID from AnyMessage.
func (fields *ChatActionFields) ParseFields(m AnyMessage) {
	fields.SenderID = uint64(m.Number(KeySenderID))
	fields.RoomID = uint64(m.Number(KeyRoomID))
}

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

func ParseErrorMessage(m AnyMessage, action Action) (ErrorMessage, error) {
	if action != ActionError {
		return ErrorMessage{}, errors.New("ParseErrorMessage: invalid action")
	}
	msg := ErrorMessage{}
	msg.ActionName = action
	msg.ErrorMsg = m.String("error")
	msg.Cause = AnyMessage(m.Object("cause"))
	return msg, nil
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

// == ChatMessage related ActionMessages ==

// ChatMessage is chat message which is recieved from a browser-side
// client and sends to other clients in the same room.
// it implements ChatActionMessage interface.
type ChatMessage struct {
	ChatActionFields
	ID      uint64 `json:"id,omitempty"` // used only server->client
	Content string `json:"content,omitempty"`
}

func ParseChatMessage(m AnyMessage, action Action) (ChatMessage, error) {
	if action != ActionChatMessage {
		return ChatMessage{}, errors.New("ParseChatMessage: invalid action")
	}
	cm := ChatMessage{}
	cm.ActionName = action
	cm.Content = m.String("content")
	cm.ChatActionFields.ParseFields(m)
	return cm, nil
}

// ReadMessage indicates notification which some chat messages are read by
// any user.
// it implements ChatActionMessage interface.
type ReadMessage struct {
	ChatActionFields
	MessageIDs []uint64 `json:"message_ids"`
}

func ParseReadMessage(m AnyMessage, action Action) (ReadMessage, error) {
	if action != ActionReadMessage {
		return ReadMessage{}, errors.New("ParseReadMessage: invalid action")
	}
	rm := ReadMessage{}
	rm.ActionName = action
	rm.ChatActionFields.ParseFields(m)
	rm.MessageIDs = m.UInt64s("message_ids")
	return rm, nil
}

// TypeStart indicates user starts key typing.
// it implements ChatActionMessage interface.
type TypeStart struct {
	ChatActionFields

	// set by server and return client
	SenderName string    `json:"sender_name,omitempty"`
	StartAt    time.Time `json:"start_at,omitempty"`
}

func ParseTypeStart(m AnyMessage, action Action) (TypeStart, error) {
	if action != ActionTypeStart {
		return TypeStart{}, errors.New("ParseTypeStart: invalid action")
	}
	ts := TypeStart{}
	ts.ActionName = action
	ts.ChatActionFields.ParseFields(m)
	return ts, nil
}

// TypeEnd indicates user ends key typing.
// it implements ChatActionMessage interface.
type TypeEnd struct {
	ChatActionFields

	// set by server and return client
	SenderName string    `json:"sender_name,omitempty"`
	EndAt      time.Time `json:"end_at,omitempty"`
}

func ParseTypeEnd(m AnyMessage, action Action) (TypeEnd, error) {
	if action != ActionTypeEnd {
		return TypeEnd{}, errors.New("ParseTypeEnd: invalid action")
	}
	te := TypeEnd{}
	te.ActionName = action
	te.ChatActionFields.ParseFields(m)
	return te, nil
}

// CreateRoom indicates action for a request for creating room.
// it implements ActionMessage interface.
type CreateRoom struct {
	EmbdFields

	SenderID      uint64   `json:"sender_id"`
	RoomName      string   `json:"room_name"`
	RoomMemberIDs []uint64 `json:"room_member_ids"`
}

func ParseCreateRoom(m AnyMessage, action Action) (CreateRoom, error) {
	if action != ActionCreateRoom {
		return CreateRoom{}, errors.New("CreateRoom: invalid action")
	}
	cr := CreateRoom{}
	cr.ActionName = action
	cr.SenderID = uint64(m.Number("sender_id"))
	cr.RoomName = m.String("room_name")
	cr.RoomMemberIDs = m.UInt64s("room_member_ids")
	return cr, nil
}

// DeleteRoom indicates action for a request for deleting room.
// it implements ActionMessage interface.
type DeleteRoom struct {
	EmbdFields

	SenderID uint64 `json:"sender_id"`
	RoomID   uint64 `json:"room_id"`
}

func ParseDeleteRoom(m AnyMessage, action Action) (DeleteRoom, error) {
	if action != ActionDeleteRoom {
		return DeleteRoom{}, errors.New("DeleteRoom: invalid action")
	}
	dr := DeleteRoom{}
	dr.ActionName = action
	dr.SenderID = uint64(m.Number("sender_id"))
	dr.RoomID = uint64(m.Number("room_id"))
	return dr, nil
}
