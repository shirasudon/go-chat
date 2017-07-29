package entity

type Message struct {
	ID       uint   `json:"id,omitempty"`
	Content  string `json:"content,omitempty"`
	SenderID int64  `json:"senderID,omitempty"`
	RoomID   int64  `json:"roomID,omitempty"`
}

type MessageRepository interface {
	GetFromRoom(roomID int64, n int) ([]Message, error)
	AddToRoom(roomID int64, m Message) error
}

type MessageRepositoryStub struct {
	messages []Message
}

func NewMessageRepositoryStub() *MessageRepositoryStub {
	return &MessageRepositoryStub{
		messages: make([]Message, 0, 100),
	}
}

func (repo *MessageRepositoryStub) AddToRoom(roomID, m Message) error {
	repo.messages = append(repo.messages, m)
	return nil
}

func (repo *MessageRepositoryStub) GetFromRoom(roomID int64, n int) ([]Message, error) {
	if n > len(repo.messages) {
		n = len(repo.messages)
	}
	return repo.messages[:n], nil
}
