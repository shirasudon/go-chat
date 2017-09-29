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

func (s ChatService) FindRoomRelation(ctx context.Context, roomID uint64) (entity.RoomRelation, error) {
	return s.repos.RoomRelations().Find(ctx, roomID)
}

func (s ChatService) FindUserRelation(ctx context.Context, userID uint64) (entity.UserRelation, error) {
	return s.repos.UserRelations().Find(ctx, userID)
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