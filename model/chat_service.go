package model

import (
	"context"

	"github.com/shirasudon/go-chat/entity"
	"github.com/shirasudon/go-chat/model/action"
)

// ChatService provides the usecases for
// chat messaging application.
type ChatService struct {
	repos entity.Repositories
	msgs  entity.MessageRepository
	users entity.UserRepository
	rooms entity.RoomRepository
}

func NewChatService(repos entity.Repositories) *ChatService {
	return &ChatService{
		repos: repos,
		msgs:  repos.Messages(),
		users: repos.Users(),
		rooms: repos.Rooms(),
	}
}

// Find friend users related with specified user id.
// It returns error if not found.
func (s ChatService) FindUserFriends(ctx context.Context, userID uint64) ([]entity.User, error) {
	return s.users.FindAllByUserID(ctx, userID)
}

// Find rooms related with specified user id.
// It returns error if not found.
func (s ChatService) FindUserRooms(ctx context.Context, userID uint64) ([]entity.Room, error) {
	return s.rooms.FindAllByUserID(ctx, userID)
}

// UserRelation is the relationship owned by specified UserID.
type UserRelation struct {
	UserID  uint64
	Friends []entity.User
	Rooms   []entity.Room
}

// Find both of friends and rooms related with specified user id.
// It returns error if not found.
func (s ChatService) FindUserRelation(ctx context.Context, userID uint64) (UserRelation, error) {
	users, err1 := s.users.FindAllByUserID(ctx, userID)
	if err1 != nil {
		return UserRelation{}, err1
	}
	rooms, err := s.rooms.FindAllByUserID(ctx, userID)

	return UserRelation{
		UserID:  userID,
		Friends: users,
		Rooms:   rooms,
	}, err
}

// Post the message to the specified room.
// It returns posted message id and nil or error
// which indicates the message can not be posted.
func (s ChatService) PostRoomMessage(ctx context.Context, m action.ChatMessage) (msgID uint64, err error) {
	msgID, err = s.msgs.Add(ctx, entity.Message{
		Content: m.Content,
		UserID:  m.SenderID,
		RoomID:  m.RoomID,
	})
	return
}

// Mark the message is read by the specified user.
// It returns error when the message can not be marked to read.
func (s ChatService) ReadRoomMessage(ctx context.Context, m action.ReadMessage) error {
	return s.msgs.ReadMessage(ctx, m.RoomID, m.SenderID, m.MessageIDs)
}
