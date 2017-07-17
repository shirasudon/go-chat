package chat

type MessageRepository interface {
	Put(Message) error
	Get() ([]Message, error)
}

type MessageRepositoryStub struct {
	messages []Message
}

func NewMessageRepositoryStub() *MessageRepositoryStub {
	return &MessageRepositoryStub{
		messages: make([]Message, 0, 100),
	}
}

func (repo *MessageRepositoryStub) Put(m Message) error {
	repo.messages = append(repo.messages, m)
	return nil
}

func (repo *MessageRepositoryStub) Get() ([]Message, error) {
	return repo.messages, nil
}
