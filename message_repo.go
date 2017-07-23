package chat

type MessageRepository interface {
	Put(ChatMessage) (registeredID uint, err error)
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

func (repo *MessageRepositoryStub) Put(m ChatMessage) (registeredID uint, err error) {
	repo.messages = append(repo.messages, m)
	return uint(len(repo.messages)) - 1, nil
}

func (repo *MessageRepositoryStub) Get() ([]ChatMessage, error) {
	return repo.messages, nil
}
