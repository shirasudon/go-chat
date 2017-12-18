package chat

import (
	"context"
	"fmt"
	"time"

	"github.com/shirasudon/go-chat/chat/action"
	"github.com/shirasudon/go-chat/chat/queried"
	"github.com/shirasudon/go-chat/domain/event"
)

//go:generate mockgen -destination=../internal/mocks/mock_query_service.go -package=mocks github.com/shirasudon/go-chat/chat QueryService

// QueryService is the interface for the querying the information from backend datastore.
type QueryService interface {
	// It finds authenticated user profile matched with given user name and password.
	// It returns queried user profile and nil when the user is found in the data-store, or
	// returns nil and NotFoundError when the user is not found.
	FindUserByNameAndPassword(ctx context.Context, name, password string) (*queried.AuthUser, error)

	// Find the relational information of user specified by userID.
	// It returns queried result and nil, or nil and NotFoundError if the information is not found.
	FindUserRelation(ctx context.Context, userID uint64) (*queried.UserRelation, error)

	// Find the room information specified by roomID with userID.
	// It returns queried result and nil, or nil and NotFoundError if the information is not found.
	FindRoomInfo(ctx context.Context, userID, roomID uint64) (*queried.RoomInfo, error)

	// Find the messages belonging to the room specified by QueryRoomMessages with userID.
	// It returns queried messages and nil, or nil and InfraError if infrastructure raise some errors.
	FindRoomMessages(ctx context.Context, userID uint64, q action.QueryRoomMessages) (*queried.RoomMessages, error)

	// Find the unread messages belonging to the room specified by QueryUnreadRoomMessages with userID.
	// It returns queried messages and nil, or nil and InfraError if infrastructure raise some errors.
	FindUnreadRoomMessages(ctx context.Context, userID uint64, q action.QueryUnreadRoomMessages) (*queried.UnreadRoomMessages, error)
}

// TODO cache feature.

// QueryServiceImpl implements QueryService interface.
type QueryServiceImpl struct {
	users UserQueryer
	rooms RoomQueryer
	msgs  MessageQueryer

	events EventQueryer
}

func NewQueryServiceImpl(qs *Queryers) *QueryServiceImpl {
	if qs == nil {
		panic("nil Queryers")
	}
	return &QueryServiceImpl{
		users:  qs.UserQueryer,
		rooms:  qs.RoomQueryer,
		msgs:   qs.MessageQueryer,
		events: qs.EventQueryer,
	}
}

// TODO event permittion in for user id,
func (s *QueryServiceImpl) FindEventsByTimeCursor(ctx context.Context, after time.Time, limit int) ([]event.Event, error) {
	evs, err := s.events.FindAllByTimeCursor(ctx, after, limit)
	if err != nil && IsNotFoundError(err) {
		return []event.Event{}, nil
	}
	return evs, err
}

// Find user profile matched with user name and password.
// It returns queried user profile and nil when found in the data-store.
// It returns nil and error when the user is not found.
func (s *QueryServiceImpl) FindUserByNameAndPassword(ctx context.Context, name, password string) (*queried.AuthUser, error) {
	user, err := s.users.FindByNameAndPassword(ctx, name, password)
	// TODO cache?
	return user, err
}

// Find abstract information associated with the User.
// It returns queried result and error if the information is not found.
func (s *QueryServiceImpl) FindUserRelation(ctx context.Context, userID uint64) (*queried.UserRelation, error) {
	relation, err := s.users.FindUserRelation(ctx, userID)
	// TODO cache?
	return relation, err
}

// Find detailed room information specified by room ID.
// It also requires userID to query the information which
// can be permmited to the user.
// It returns queried room information and error if not found.
func (s *QueryServiceImpl) FindRoomInfo(ctx context.Context, userID, roomID uint64) (*queried.RoomInfo, error) {
	info, err := s.rooms.FindRoomInfo(ctx, userID, roomID)
	// TODO cache?
	return info, err
}

const (
	MaxRoomMessagesLimit = 50
)

// Find messages from specified room.
// It returns error if infrastructure raise some errors.
func (s *QueryServiceImpl) FindRoomMessages(ctx context.Context, userID uint64, q action.QueryRoomMessages) (*queried.RoomMessages, error) {
	// check query paramnter
	if q.Limit > MaxRoomMessagesLimit {
		q.Limit = MaxRoomMessagesLimit
	} else if q.Limit < 0 {
		q.Limit = 0
	}

	if q.Before == (time.Time{}) {
		q.Before = time.Now()
	}

	// TODO create specific queried data, messages associated with room ID and user ID,
	// to remove domain logic in the QueryService.
	r, err := s.rooms.Find(ctx, q.RoomID)
	if err != nil {
		return nil, err
	}
	u, err := s.users.Find(ctx, userID)
	if err != nil {
		return nil, err
	}
	if !r.HasMember(u) {
		return nil, fmt.Errorf("can not get the messages from room(id=%d) by not a room member user(id=%d)", q.RoomID, userID)
	}

	msgs, err := s.msgs.FindRoomMessagesOrderByLatest(ctx, q.RoomID, q.Before, q.Limit)
	if err != nil {
		if IsNotFoundError(err) {
			res := queried.EmptyRoomMessages
			res.RoomID = q.RoomID
			return &res, nil
		}
		return nil, err
	}

	// TODO move to infrastructure and just return QueryRoomMessages.

	roomMsgs := &queried.RoomMessages{
		RoomID: q.RoomID,
	}
	roomMsgs.Cursor.Current = q.Before
	if last := len(msgs) - 1; last >= 0 {
		roomMsgs.Cursor.Next = msgs[last].CreatedAt
	} else {
		roomMsgs.Cursor.Next = q.Before
	}

	qMsgs := make([]queried.Message, 0, len(msgs))
	for _, m := range msgs {
		qm := queried.Message{
			MessageID: m.ID,
			UserID:    m.UserID,
			Content:   m.Content,
			CreatedAt: m.CreatedAt,
		}
		qMsgs = append(qMsgs, qm)
	}
	roomMsgs.Msgs = qMsgs

	return roomMsgs, nil
}

// Find unread messages from specified room.
// It returns error if infrastructure raise some errors.
func (s *QueryServiceImpl) FindUnreadRoomMessages(ctx context.Context, userID uint64, q action.QueryUnreadRoomMessages) (*queried.UnreadRoomMessages, error) {
	// check existance of user and room.
	if _, err := s.users.Find(context.Background(), userID); err != nil {
		return nil, err
	}
	if _, err := s.rooms.Find(context.Background(), q.RoomID); err != nil {
		return nil, err
	}

	msgs, err := s.msgs.FindUnreadRoomMessages(ctx, userID, q.RoomID, q.Limit)
	if err != nil && IsNotFoundError(err) {
		// return empty result because room exists but message is not yet.
		res := queried.EmptyUnreadRoomMessages
		res.RoomID = q.RoomID
		return &res, nil
	}
	// TODO cache
	return msgs, err
}
