package stub

import (
	"context"
	"errors"
	"time"

	"github.com/mzki/chat/entity"
)

type MessageRepository struct {
	messages []entity.Message
}

func newMessageRepository() *MessageRepository {
	return &MessageRepository{
		messages: make([]entity.Message, 0, 100),
	}
}

func (repo *MessageRepository) LatestRoomMessages(ctx context.Context, roomID uint64, n int) ([]entity.Message, error) {
	l := len(repo.messages)
	if offset := l - n; offset > 0 {
		return repo.messages[offset : offset+n], nil
	}
	return repo.messages[:], nil
}

func (repo *MessageRepository) PreviousRoomMessages(ctx context.Context, offset entity.Message, n int) ([]entity.Message, error) {
	var offsetIdx int
	for i, m := range repo.messages {
		if m.ID > offset.ID {
			offsetIdx = i
			break
		}
	}
	if offsetIdx == 0 {
		return nil, errors.New("not found")
	}
	if offsetIdx < n {
		return repo.messages[:offsetIdx], nil
	}
	return repo.messages[offsetIdx-n : offsetIdx], nil
}

func (repo *MessageRepository) Save(m entity.Message) (uint64, error) {
	m.ID = uint64(len(repo.messages))
	m.CreatedAt = time.Now()
	repo.messages = append(repo.messages)
	return m.ID, nil
}
