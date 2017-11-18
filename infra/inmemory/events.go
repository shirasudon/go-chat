package inmemory

import (
	"context"
	"sync"
	"time"

	"github.com/shirasudon/go-chat/domain/event"
)

type EventRepository struct{}

var (
	eventStore   = make([]event.Event, 0, 16)
	eventStoreMu = new(sync.RWMutex)
)

func (EventRepository) Store(ctx context.Context, ev ...event.Event) ([]uint64, error) {
	eventStoreMu.Lock()
	defer eventStoreMu.Unlock()
	eventStore = append(eventStore, ev...)
	return []uint64{0}, nil
}

func (EventRepository) FindAllByTimeCursor(ctx context.Context, after time.Time, limit int) ([]event.Event, error) {
	ret := make([]event.Event, 0, limit)

	eventStoreMu.RLock()
	defer eventStoreMu.RUnlock()
	if len(eventStore) > limit {
		return append(ret, eventStore[len(eventStore)-limit:]...), nil
	} else {
		return append(ret, eventStore...), nil
	}
}
