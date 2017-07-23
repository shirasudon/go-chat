package chat

type MessageRepository interface {
	Put(ChatMessage) error
	Get() ([]ChatMessage, error)
}

type MessageRepositoryStub struct {
	messages []ChatMessage
}

func NewMessageRepositoryStub() *MessageRepositoryStub {
	return &MessageRepositoryStub{
		messages: make([]ChatMessage, 0, 100),
	}
}

func (repo *MessageRepositoryStub) Put(m ChatMessage) error {
	repo.messages = append(repo.messages, m)
	return nil
}

func (repo *MessageRepositoryStub) Get() ([]ChatMessage, error) {
	return repo.messages, nil
}
