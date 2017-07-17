package chat

type Message struct {
	Content  string `json:"content",omitempty`
	Sender   string `json:"sender",omitempty`
	Receiver string `json:"receiver",omitempty`
}

func (m Message) String() string {
	return m.Sender + ": " + m.Content
}
