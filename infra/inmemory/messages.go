package inmemory

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/shirasudon/go-chat/chat"
	"github.com/shirasudon/go-chat/chat/queried"
	"github.com/shirasudon/go-chat/domain"
	"github.com/shirasudon/go-chat/domain/event"
)

var (
	messageMapMu *sync.RWMutex = new(sync.RWMutex)

	messageMap = map[uint64]domain.Message{
		1: {
			ID:        1,
			Content:   "hello!",
			CreatedAt: time.Now().Add(-10 * time.Millisecond),
			UserID:    2,
			RoomID:    2,
		},
	}

	messageCounter uint64 = uint64(len(messageMap))

	// key: user-room ID, value: message id map
	userAndRoomIDToUnreadMessageIDs = map[userAndRoomID]map[uint64]bool{}
)

type userAndRoomID struct {
	UserID uint64
	RoomID uint64
}

func errMsgNotFound(msgID uint64) *chat.NotFoundError {
	return chat.NewNotFoundError("message (id=%v) is not found")
}

type MessageRepository struct {
	domain.EmptyTxBeginner
	pubsub chat.Pubsub
}

func NewMessageRepository(pubsub chat.Pubsub) *MessageRepository {
	return &MessageRepository{
		pubsub: pubsub,
	}
}

// It runs infinite loop for updating query data by domain events.
// if context is canceled, the infinite loop quits.
// It must be called to be updated to latest query data.
func (repo *MessageRepository) UpdatingService(ctx context.Context) {
	evCh := repo.pubsub.Sub(
		event.TypeMessageCreated,
		event.TypeMessageReadByUser,
	)
	for {
		select {
		case ev, ok := <-evCh:
			if !ok {
				return
			}
			if ev, ok := ev.(event.Event); ok {
				repo.updateByEvent(ev)
			}
		case <-ctx.Done():
			return
		}
	}
}

func (repo *MessageRepository) updateByEvent(ev event.Event) {
	switch ev := ev.(type) {
	case event.MessageCreated:
		messageMapMu.Lock()
		defer messageMapMu.Unlock()

		key := userAndRoomID{ev.CreatedBy, ev.RoomID}
		msgIDs, ok := userAndRoomIDToUnreadMessageIDs[key]
		if !ok {
			msgIDs = make(map[uint64]bool)
			userAndRoomIDToUnreadMessageIDs[key] = msgIDs
		}
		msgIDs[ev.MessageID] = true

	case event.MessageReadByUser:
		messageMapMu.Lock()
		defer messageMapMu.Unlock()

		if msgIDs, ok := userAndRoomIDToUnreadMessageIDs[userAndRoomID{ev.UserID, ev.RoomID}]; ok {
			delete(msgIDs, ev.MessageID)
		}
	}
}

func (repo *MessageRepository) Find(ctx context.Context, msgID uint64) (domain.Message, error) {
	m, ok := messageMap[msgID]
	if ok {
		return m, nil
	}
	return domain.Message{}, errMsgNotFound(msgID)
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

	if len(msgs) > limit {
		msgs = msgs[:limit]
	}
	return msgs, nil
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

func (repo *MessageRepository) FindUnreadRoomMessages(ctx context.Context, userID, roomID uint64, limit int) (*queried.UnreadRoomMessages, error) {
	messageMapMu.RLock()
	defer messageMapMu.RUnlock()

	unreadIDs, ok := userAndRoomIDToUnreadMessageIDs[userAndRoomID{userID, roomID}]
	if !ok {
		return nil, chat.NewNotFoundError("user (id=%v) has no unread messsages for the room (id=%v)", userID, roomID)
	}

	unreadMsgs := make([]queried.Message, 0, len(unreadIDs))
	for id, _ := range unreadIDs {
		if m, ok := messageMap[id]; ok {
			qm := queried.Message{
				MessageID: m.ID,
				UserID:    m.UserID,
				Content:   m.Content,
				CreatedAt: m.CreatedAt,
			}
			unreadMsgs = append(unreadMsgs, qm)
		}
	}

	return &queried.UnreadRoomMessages{
		RoomID:   roomID,
		Msgs:     unreadMsgs,
		MsgsSize: len(unreadMsgs),
	}, nil
}
