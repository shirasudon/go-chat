package chat

import (
	"context"
	"fmt"
	"time"

	"github.com/shirasudon/go-chat/chat/action"
	"github.com/shirasudon/go-chat/chat/queried"
	"github.com/shirasudon/go-chat/domain/event"
)

// TODO cache feature.

// QueryService queries the action message data
// from the datastores.
type QueryService struct {
	users UserQueryer
	rooms RoomQueryer
	msgs  MessageQueryer

	events EventQueryer
}

func NewQueryService(qs *Queryers) *QueryService {
	if qs == nil {
		panic("nil Queryers")
	}
	return &QueryService{
		users:  qs.UserQueryer,
		rooms:  qs.RoomQueryer,
		msgs:   qs.MessageQueryer,
		events: qs.EventQueryer,
	}
}

func (s *QueryService) FindEventsByTimeCursor(ctx context.Context, after time.Time, limit int) ([]event.Event, error) {
	return s.events.FindAllByTimeCursor(ctx, after, limit)
}

// Find rooms related with specified user id.
// It returns error if not found.
func (s *QueryService) FindUserRooms(ctx context.Context, userID uint64) ([]queried.Room, error) {
	rooms, err := s.rooms.FindAllByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	userRooms := make([]queried.Room, 0, len(rooms))
	for _, r := range rooms {
		ur := queried.Room{
			RoomName: r.Name,
			OwnerID:  userID,
		}
		// TODO get user information
		// ur.Members = ...
		userRooms = append(userRooms, ur)
	}
	return userRooms, nil
}

// Find abstarct information accociated with the User.
// It returns queried result and error if the information is not found.
func (s *QueryService) FindUserRelation(ctx context.Context, userID uint64) (*queried.UserRelation, error) {
	relation, err := s.users.FindUserRelation(ctx, userID)
	// TODO cache?
	return relation, err
}

// Find messages from specified room.
// It returns error if infrastructure raise some errors.
func (s *QueryService) FindRoomMessages(ctx context.Context, userID uint64, q action.QueryRoomMessages) (*queried.RoomMessages, error) {
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
			Content:   m.Content,
			CreatedAt: m.CreatedAt,
		}
		qMsgs = append(qMsgs, qm)
	}
	roomMsgs.Msgs = qMsgs

	return roomMsgs, nil
}
