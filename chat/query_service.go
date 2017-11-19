package chat

import (
	"context"
	"fmt"
	"time"

	"github.com/shirasudon/go-chat/chat/action"
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

type QueriedUserRoom struct {
	RoomName string `json:"room_name"`
	OwnerID  uint64 `json:"owner_id"`
	Members  []struct {
		UserID   uint64 `json:"user_id"`
		UserName string `json:"user_name"`
	} `json:"members"`
}

// Find rooms related with specified user id.
// It returns error if not found.
func (s *QueryService) FindUserRooms(ctx context.Context, userID uint64) ([]QueriedUserRoom, error) {
	rooms, err := s.rooms.FindAllByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	userRooms := make([]QueriedUserRoom, 0, len(rooms))
	for _, r := range rooms {
		ur := QueriedUserRoom{
			RoomName: r.Name,
			OwnerID:  userID,
		}
		// TODO get user information
		// ur.Members = ...
		userRooms = append(userRooms, ur)
	}
	return userRooms, nil
}

// QueriedUserRelation is the abstarct information associated with specified User.
type QueriedUserRelation struct {
	UserID   uint64 `json:"user_id"`
	UserName string `json:"user_name"`
	// TODO first name, last name

	Friends []UserFriend `json:"friends"`

	Rooms []UserRoom `json:"rooms"`
}

type UserFriend struct {
	UserID   uint64 `json:"user_id"`
	UserName string `json:"user_name"`
}

type UserRoom struct {
	RoomID   uint64 `json:"room_id"`
	RoomName string `json:"room_name"`
}

// Find abstarct information accociated with the User.
// It returns queried result and error if the information is not found.
func (s *QueryService) FindUserRelation(ctx context.Context, userID uint64) (*QueriedUserRelation, error) {
	relation, err := s.users.FindUserRelation(ctx, userID)
	// TODO cache?
	return relation, err
}

type QueriedRoomMessages struct {
	RoomID uint64 `json:"room_id"`

	Msgs []QueriedMessage `json:"messages"`

	Cursor struct {
		Current time.Time `json:"current"`
		Next    time.Time `json:"next"`
	} `json:"cursor"`
}

type QueriedMessage struct {
	MessageID uint64    `json:"message_id"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

// Find messages from specified room.
// It returns error if infrastructure raise some errors.
func (s *QueryService) FindRoomMessages(ctx context.Context, userID uint64, q action.QueryRoomMessages) (*QueriedRoomMessages, error) {
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

	roomMsgs := &QueriedRoomMessages{
		RoomID: q.RoomID,
	}
	roomMsgs.Cursor.Current = q.Before
	if last := len(msgs) - 1; last >= 0 {
		roomMsgs.Cursor.Next = msgs[last].CreatedAt
	} else {
		roomMsgs.Cursor.Next = q.Before
	}

	qMsgs := make([]QueriedMessage, 0, len(msgs))
	for _, m := range msgs {
		qm := QueriedMessage{
			MessageID: m.ID,
			Content:   m.Content,
			CreatedAt: m.CreatedAt,
		}
		qMsgs = append(qMsgs, qm)
	}
	roomMsgs.Msgs = qMsgs

	return roomMsgs, nil
}
