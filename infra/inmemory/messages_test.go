package inmemory

import (
	"context"
	"testing"
	"time"

	"github.com/shirasudon/go-chat/domain"
	"github.com/shirasudon/go-chat/domain/event"
	"github.com/shirasudon/go-chat/infra/pubsub"
)

var (
	globalPubsub      = pubsub.New()
	messageRepository = NewMessageRepository(globalPubsub)
)

func TestMessageRepoUpdatingService(t *testing.T) {
	// with timeout to quit correctly
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	// run service for updating query data
	go messageRepository.UpdatingService(ctx)

	ch := globalPubsub.Sub(event.TypeMessageCreated)

	const (
		TargetRoomID = 1
		TargetUserID = 1
		Content      = "none"
	)

	for i := 0; i < 10; i++ {
		id, _ := messageRepository.Store(ctx, domain.Message{Content: Content})
		globalPubsub.Pub(event.MessageCreated{MessageID: id, CreatedBy: TargetUserID, RoomID: TargetRoomID})
		select {
		case <-ch:
			continue
		case <-ctx.Done():
			t.Fatal("timeout")
		}
	}
}

func TestMessageRepoStore(t *testing.T) {
	const Content = "hello"
	currentID := uint64(len(messageMap))
	id, err := messageRepository.Store(context.Background(), domain.Message{Content: Content})
	if err != nil {
		t.Fatal(err)
	}

	if id != currentID+1 {
		t.Fatalf("different created message id, expect: %v, got: %v", currentID+1, id)
	}

	stored, ok := messageMap[id]
	if !ok {
		t.Fatal("message created but not stored")
	}
	if stored.ID != id {
		t.Errorf("different message id in the datastore, expect: %v, got: %v", id, stored.ID)
	}
	if stored.Content != Content {
		t.Errorf("different message content in the datastore, expect: %v, got: %v", Content, stored.Content)
	}
}

func TestMessageRepoFind(t *testing.T) {
	// case1: found message
	const Content = "hello1"
	id, err := messageRepository.Store(context.Background(), domain.Message{Content: Content})
	if err != nil {
		t.Fatal(err)
	}

	m, err := messageRepository.Find(context.Background(), id)
	if err != nil {
		t.Fatalf("can not find any message: %v", err)
	}
	if m.Content != Content {
		t.Errorf("different message content, expect: %v, got: %v", Content, m.Content)
	}

	// case2: not found message
	const NotFoundID = 99999
	if _, err := messageRepository.Find(context.Background(), NotFoundID); err == nil {
		t.Fatal("find by not found id but no error")
	}
}

func TestMessageRepoFindRoomMessagesOrderByLatest(t *testing.T) {
	// case1: limit 0
	ms, err := messageRepository.FindRoomMessagesOrderByLatest(
		context.Background(),
		0,
		time.Time{},
		0,
	)
	if err != nil {
		t.Fatalf("limit 0 but error returned: %v", err)
	}
	if len(ms) != 0 {
		t.Errorf("limit 0 but some message returned: %v", ms)
	}

	// case2: find success
	const FoundRoomID = 900
	beforeCreated := time.Now()
	for _, m := range []domain.Message{
		{RoomID: FoundRoomID, Content: "1", CreatedAt: beforeCreated.Add(10 * time.Millisecond)},
		{RoomID: FoundRoomID, Content: "2", CreatedAt: beforeCreated.Add(11 * time.Millisecond)},
	} {
		_, err := messageRepository.Store(context.Background(), m)
		if err != nil {
			t.Fatalf("storing error with %v: %v", m, err)
		}
	}
	afterCreated := time.Now()

	ms, err = messageRepository.FindRoomMessagesOrderByLatest(
		context.Background(),
		FoundRoomID,
		afterCreated,
		10,
	)
	if err != nil {
		t.Fatal(err)
	}
	if len(ms) != 2 {
		t.Fatalf("different room messages size, expect: %v, got: %v", 2, len(ms))
	}
	if ms[0].CreatedAt.Before(ms[1].CreatedAt) {
		t.Errorf("different order for the room messages")
	}
}

func TestMessageRepoRemoveAllByRoomID(t *testing.T) {
	// case 1: remove targets are found
	const FoundRoomID = 900
	if err := messageRepository.RemoveAllByRoomID(context.Background(), FoundRoomID); err != nil {
		t.Fatal(err)
	}

	ms, err := messageRepository.FindRoomMessagesOrderByLatest(
		context.Background(), FoundRoomID, time.Now(), 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(ms) != 0 {
		t.Errorf("removed messages are found")
	}

	// case 2: remove targets are not found
	const NotFoundRoomID = FoundRoomID + 10
	if err := messageRepository.RemoveAllByRoomID(context.Background(), NotFoundRoomID); err != nil {
		t.Fatal("remove by not found room id but error occured")
	}
}

func TestMessageRepoFindUnreadRoomMessages(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Millisecond)
	defer cancel()

	const (
		TargetRoomID  = 999
		TargetUserID  = 1
		ContentUnread = "hello1"
		ContentRead   = "hello2"
	)

	messageRepository.Store(ctx, domain.Message{Content: ContentUnread, RoomID: TargetRoomID})
	id, _ := messageRepository.Store(ctx, domain.Message{Content: ContentRead, RoomID: TargetRoomID})
	ev := event.MessageReadByUser{MessageID: id, UserID: TargetUserID, RoomID: TargetRoomID}
	messageRepository.updateByEvent(ev)

	unreads, err := messageRepository.FindUnreadRoomMessages(ctx, TargetUserID, TargetRoomID, 1)
	if err != nil {
		t.Fatal(err)
	}

	if unreads.RoomID != TargetRoomID {
		t.Errorf("different RoomID, expect: %v, got: %v", TargetRoomID, unreads.RoomID)
	}
	if unreads.MsgsSize != 1 {
		t.Errorf("different unread messages size, expect: %v, got: %v", 1, unreads.MsgsSize)
	}
	if unreads.Msgs[0].Content != ContentUnread {
		t.Errorf("different queried messages content, expect: %v, got: %v", ContentUnread, unreads.Msgs[0].Content)
	}
}
