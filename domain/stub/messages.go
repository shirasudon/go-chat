package stub

import (
	"context"
	"errors"
	"time"

	"github.com/shirasudon/go-chat/domain"
)

type MessageRepository struct {
	domain.EmptyTxBeginner

	messages []domain.Message
}

func newMessageRepository() *MessageRepository {
	return &MessageRepository{
		messages: make([]domain.Message, 0, 100),
	}
}

func (repo *MessageRepository) LatestRoomMessages(ctx context.Context, roomID uint64, n int) ([]domain.Message, error) {
	l := len(repo.messages)
	if offset := l - n; offset > 0 {
		return repo.messages[offset : offset+n], nil
	}
	return repo.messages[:], nil
}

func (repo *MessageRepository) PreviousRoomMessages(ctx context.Context, offset domain.Message, n int) ([]domain.Message, error) {
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

func (repo *MessageRepository) Add(ctx context.Context, m domain.Message) (uint64, error) {
	m.ID = uint64(len(repo.messages))
	m.CreatedAt = time.Now()
	repo.messages = append(repo.messages)
	return m.ID, nil
}

func (repo *MessageRepository) ReadMessage(ctx context.Context, roomID, userID uint64, messageIDs []uint64) error {
	// No-op
	return nil
}
