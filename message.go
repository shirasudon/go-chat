package chat

// message is a payload through the websocket.
type Message map[string]interface{}

// Action indicates a action type for the JSON data schema.
type Action string

// key for the action field in JSON.
const KeyAction = "action"

const (
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

type ChatMessage struct {
	embdFields
	Content  string `json:"content,omitempty"`
	SenderID int64  `json:"senderID,omitempty"`
	RoomID   int64  `json:"roomID,omitempty"`
}

func ParseChatMessage(m Message, action Action) ChatMessage {
	cm := ChatMessage{}
	cm.Action = action
	cm.Content, _ = m["content"].(string)
	cm.SenderID, _ = m["senderID"].(int64)
	cm.RoomID, _ = m["roomID"].(int64)
	return cm
}

type UserConnect struct {
	embdFields
	UserID int `json:"userID,omitempty"`
}
