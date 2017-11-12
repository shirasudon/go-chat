package chat

import (
	"context"
	"time"

	"github.com/shirasudon/go-chat/chat/action"
)

// TODO cache feature.

// QueryService queries the action message data
// from the datastores.
type QueryService struct {
	users UserQueryer
	rooms RoomQueryer
	msgs  MessageQueryer
}

func NewQueryService(qs *Queryers) *QueryService {
	if qs == nil {
		panic("nil Queryers")
	}
	return &QueryService{
		users: qs.UserQueryer,
		rooms: qs.RoomQueryer,
		msgs:  qs.MessageQueryer,
	}
}

// Find friend users related with specified user id.
// It returns error if not found.
// func (s *QueryService) FindUserFriends(ctx context.Context, userID uint64) ([]domain.User, error) {
// 	return s.users.FindAllByUserID(ctx, userID)
// }

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

// UserRelation is the relationship owned by specified UserID.
// type UserRelation struct {
// 	UserID  uint64
// 	Friends []domain.User
// 	Rooms   []domain.Room
// }

// Find both of friends and rooms related with specified user id.
// It returns error if not found.
// func (s QueryService) FFindUserRelation(ctx context.Context, userID uint64) (UserRelation, error) {
// 	users, err1 := s.users.FindAllByUserID(ctx, userID)
// 	if err1 != nil {
// 		return UserRelation{}, err1
// 	}
// 	rooms, err := s.rooms.FindAllByUserID(ctx, userID)
//
// 	return UserRelation{
// 		UserID:  userID,
// 		Friends: users,
// 		Rooms:   rooms,
// 	}, err
// }

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
func (s *QueryService) FindRoomMessages(ctx context.Context, q action.QueryRoomMessages) (*QueriedRoomMessages, error) {
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
