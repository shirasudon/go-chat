package inmemory

import (
	"context"
	"sort"
	"time"

	"github.com/shirasudon/go-chat/domain"
)

var (
	messageMap = map[uint64]domain.Message{}

	messageCounter uint64 = uint64(len(messageMap))
)

type MessageRepository struct {
	domain.EmptyTxBeginner
}

func NewMessageRepository() *MessageRepository {
	return &MessageRepository{}
}

func (repo *MessageRepository) Find(ctx context.Context, msgID uint64) (domain.Message, error) {
	m, ok := messageMap[msgID]
	if ok {
		return m, nil
	}
	return domain.Message{}, ErrNotFound
}

func (repo *MessageRepository) FindAllByRoomIDOrderByLatest(ctx context.Context, roomID uint64, n int) ([]domain.Message, error) {
	if n <= 0 {
		return []domain.Message{}, nil
	}

	msgs := make([]domain.Message, 0, n)
	for _, m := range messageMap {
		if m.RoomID == roomID {
			msgs = append(msgs, m)
			if len(msgs) == n {
				break
			}
		}
	}

	sort.Slice(msgs, func(i, j int) bool { return msgs[i].CreatedAt.After(msgs[j].CreatedAt) })
	return msgs, nil
}

func (repo *MessageRepository) FindRoomMessagesOrderByLatest(ctx context.Context, roomID uint64, before time.Time, limit int) ([]domain.Message, error) {
	if limit <= 0 {
		return []domain.Message{}, nil
	}

	msgs := make([]domain.Message, 0, limit)
	for _, m := range messageMap {
		if m.RoomID == roomID && m.CreatedAt.Before(before) {
			msgs = append(msgs, m)
		}
	}

	sort.Slice(msgs, func(i, j int) bool { return msgs[i].CreatedAt.After(msgs[j].CreatedAt) })

	return msgs[:limit], nil
}

func (repo *MessageRepository) Store(ctx context.Context, m domain.Message) (uint64, error) {
	// TODO create or update
	messageCounter += 1
	m.ID = messageCounter
	m.CreatedAt = time.Now()
	messageMap[m.ID] = m

	return m.ID, nil
}

func (repo *MessageRepository) RemoveAllByRoomID(ctx context.Context, roomID uint64) error {
	for id, m := range messageMap {
		if m.RoomID == roomID {
			delete(messageMap, id)
		}
	}
	return nil
}
