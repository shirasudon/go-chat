package model

import (
	"context"

	"github.com/shirasudon/go-chat/domain"
	"github.com/shirasudon/go-chat/model/action"
)

// ChatCommandService provides the usecases for
// creating/updating/editing/deleting the application
// data.
type ChatCommandService struct {
	msgs   domain.MessageRepository
	users  domain.UserRepository
	rooms  domain.RoomRepository
	pubsub Pubsub
}

func NewChatCommandService(repos domain.Repositories, pubsub Pubsub) *ChatCommandService {
	return &ChatCommandService{
		msgs:   repos.Messages(),
		users:  repos.Users(),
		rooms:  repos.Rooms(),
		pubsub: pubsub,
	}
}

func withTransaction(ctx context.Context, txBeginner domain.TxBeginner, txFunc func(ctx context.Context) error) error {
	tx, err := txBeginner.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	err = txFunc(domain.SetTx(ctx, tx))
	if err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit()
}

// Post the message to the specified room.
// It returns posted message id and nil or error
// which indicates the message can not be posted.
func (s *ChatCommandService) PostRoomMessage(ctx context.Context, m action.ChatMessage) (msgID uint64, err error) {
	room, err := s.rooms.Find(ctx, m.RoomID)
	if err != nil {
		return 0, err
	}

	user, err := s.users.Find(ctx, m.SenderID)
	if err != nil {
		return 0, err
	}

	err = withTransaction(ctx, s.msgs, func(ctx context.Context) error {
		msg, msgCreated, err := domain.NewMessage(ctx, s.msgs, user, room, m.Content)
		if err != nil {
			return err
		}
		msgID = msg.ID

		s.pubsub.Pub(msgCreated)
		return nil
	})
	return msgID, err
}

// Mark the message is read by the specified user.
// It returns error when the message can not be marked to read.
func (s ChatCommandService) ReadRoomMessage(ctx context.Context, m action.ReadMessage) error {
	user, err := s.users.Find(ctx, m.SenderID)
	if err != nil {
		return err
	}

	txErr := withTransaction(ctx, s.msgs, func(ctx context.Context) error {
		// TODO implement FindAllByMessageIDs.
		msg, err := s.msgs.Find(ctx, m.MessageIDs[0])
		if err != nil {
			return err
		}

		readMsg, err := msg.ReadBy(user)
		if err != nil {
			return err
		}

		_, err = s.msgs.Store(ctx, msg)
		if err != nil {
			return err
		}

		s.pubsub.Pub(readMsg)
		return nil
	})
	return txErr
}

// ChatQueryService queries the action message data
// from the datastores.
type ChatQueryService struct {
	users domain.UserRepository
	rooms domain.RoomRepository
}

func NewChatQueryService(repos domain.Repositories) *ChatQueryService {
	return &ChatQueryService{
		users: repos.Users(),
		rooms: repos.Rooms(),
	}
}

// Find friend users related with specified user id.
// It returns error if not found.
func (s *ChatQueryService) FindUserFriends(ctx context.Context, userID uint64) ([]domain.User, error) {
	return s.users.FindAllByUserID(ctx, userID)
}

// Find rooms related with specified user id.
// It returns error if not found.
func (s *ChatQueryService) FindUserRooms(ctx context.Context, userID uint64) ([]domain.Room, error) {
	return s.rooms.FindAllByUserID(ctx, userID)
}

// UserRelation is the relationship owned by specified UserID.
type UserRelation struct {
	UserID  uint64
	Friends []domain.User
	Rooms   []domain.Room
}

// Find both of friends and rooms related with specified user id.
// It returns error if not found.
func (s ChatQueryService) FindUserRelation(ctx context.Context, userID uint64) (UserRelation, error) {
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
