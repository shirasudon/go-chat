package sqlite3

import (
	"github.com/mzki/chat/entity"
)

type MessageRepository struct{}

func (mr MessageRepository) GetFromRoom(roomID int64, n int) ([]entity.Message, error) {
	panic("not implemented")
}

func (mr MessageRepository) AddToRoom(roomID int64, m entity.Message) error {
	panic("not implemented")
}
