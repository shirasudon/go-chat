package entity

type ChatMessage struct {
	ID       uint   `json:"id,omitempty"`
	Content  string `json:"content,omitempty"`
	SenderID int64  `json:"senderID,omitempty"`
	RoomID   int64  `json:"roomID,omitempty"`
}

type ChatMessageRepository interface {
	GetFromRoom(roomID int64, n int) ([]ChatMessage, error)
	AddToRoom(roomID int64, m ChatMessage) error
}

type MessageRepositoryStub struct {
	messages []ChatMessage
}

func NewMessageRepositoryStub() *MessageRepositoryStub {
	return &MessageRepositoryStub{
		messages: make([]ChatMessage, 0, 100),
	}
}

func (repo *MessageRepositoryStub) AddToRoom(roomID, m ChatMessage) error {
	repo.messages = append(repo.messages, m)
	return nil
}

func (repo *MessageRepositoryStub) GetFromRoom(roomID int64, n int) ([]ChatMessage, error) {
	if n > len(repo.messages) {
		n = len(repo.messages)
	}
	return repo.messages[:n], nil
}
