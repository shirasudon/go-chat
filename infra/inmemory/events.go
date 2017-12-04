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
	if len(ev) == 0 {
		return []uint64{}, nil
	}

	eventStoreMu.Lock()

	insertIdx := uint64(len(eventStore)) + 1
	eventStore = append(eventStore, ev...)

	eventStoreMu.Unlock()

	ids := make([]uint64, 0, len(ev))
	for i, _ := range ev {
		ids = append(ids, insertIdx+uint64(i))
	}
	return ids, nil
}

func (EventRepository) FindAllByTimeCursor(ctx context.Context, after time.Time, limit int) ([]event.Event, error) {
	ret := make([]event.Event, 0, limit)
	if limit == 0 {
		return ret, nil
	}

	eventStoreMu.RLock()
	defer eventStoreMu.RUnlock()

	const NotFound = -99
	var startAt int = NotFound
	for i, ev := range eventStore {
		if ev.Timestamp().After(after) {
			startAt = i - 1
			break
		}
	}

	if startAt == NotFound {
		return ret, ErrNotFound
	}

	// edge case: events[0] is after given time.
	if startAt == -1 {
		startAt = 0
	}

	if startAt+limit > len(eventStore) {
		return append(ret, eventStore[startAt:len(eventStore)]...), nil
	} else {
		return append(ret, eventStore[startAt:startAt+limit]...), nil
	}
}

func (EventRepository) FindAllByStreamID(ctx context.Context, streamID event.StreamID, after time.Time, limit int) ([]event.Event, error) {
	ret := make([]event.Event, 0, limit)
	if limit == 0 {
		return ret, nil
	}

	eventStoreMu.RLock()
	defer eventStoreMu.RUnlock()

	for _, ev := range eventStore {
		if ev.StreamID() != streamID {
			continue
		}

		ret = append(ret, ev)
		if len(ret) == limit {
			break
		}
	}

	return ret, nil
}
