package chat

import (
	"context"
	"time"

	"github.com/shirasudon/go-chat/chat/action"
	"github.com/shirasudon/go-chat/domain"
	"github.com/shirasudon/go-chat/domain/event"
)

//go:generate mockgen -destination=../internal/mocks/mock_command_service.go -package=mocks github.com/shirasudon/go-chat/chat CommandService

// CommandService is the interface for sending the command
// to the chat application.
type CommandService interface {
	// It creates room specified by given actiom.CreateRoom.
	// It returns created Room's ID and InfraError if any.
	CreateRoom(ctx context.Context, m action.CreateRoom) (roomID uint64, err error)

	// It deletes room specified by given actiom message.
	// It returns deleted Room's ID and InfraError if any.
	DeleteRoom(ctx context.Context, m action.DeleteRoom) (roomID uint64, err error)

	// Mark that the room messages are read by the specified user.
	// It returns updated room ID and nil, or
	// returns InfraError when the message can not be marked to read.
	ReadRoomMessages(ctx context.Context, m action.ReadMessages) (roomID uint64, err error)

	// Post the message to the specified room.
	// It returns posted message id and nil or InfraError
	// which indicates the message can not be posted.
	PostRoomMessage(ctx context.Context, m action.ChatMessage) (msgID uint64, err error)
}

// CommandServiceImpl provides the usecases for
// creating/updating/editing/deleting the application data.
// It implements CommandService interface.
type CommandServiceImpl struct {
	msgs         domain.MessageRepository
	users        domain.UserRepository
	rooms        domain.RoomRepository
	events       event.EventRepository
	pubsub       Pubsub
	updateCancel chan struct{}
}

func NewCommandServiceImpl(repos domain.Repositories, pubsub Pubsub) *CommandServiceImpl {
	return &CommandServiceImpl{
		msgs:         repos.Messages(),
		users:        repos.Users(),
		rooms:        repos.Rooms(),
		events:       repos.Events(),
		pubsub:       pubsub,
		updateCancel: make(chan struct{}),
	}
}

// Run updating service for the domain events.
// It blocks until calling CancelUpdate() or context is done.
func (s *CommandServiceImpl) RunUpdateService(ctx context.Context) {
	roomDeleted := s.pubsub.Sub(event.TypeRoomDeleted)
	for {
		select {
		case ev, chAlived := <-roomDeleted:
			if !chAlived {
				return
			}
			deleted := ev.(event.RoomDeleted)
			err := s.msgs.RemoveAllByRoomID(ctx, deleted.RoomID)
			// TODO error handling, create ErrorEvent? or just log?
			_ = err
		case <-ctx.Done():
			return
		case <-s.updateCancel:
			return
		}
	}
}

// Stop RunUpdateService(). Multiple calling will
// occurs panic.
func (s *CommandServiceImpl) CancelUpdateService() {
	close(s.updateCancel)
}

// Do function on the context of the transaction.
// It also commits the some domain events returned from txFunc.
func (s *CommandServiceImpl) withEventTransaction(
	ctx context.Context,
	txBeginner domain.TxBeginner,
	txFunc func(ctx context.Context) ([]event.Event, error),
) error {
	return withTransaction(ctx, txBeginner, func(ctx context.Context) error {
		events, err := txFunc(ctx)
		if err != nil {
			return err
		}

		if events != nil {
			_, err := s.events.Store(ctx, events...)
			if err != nil {
				return err
			}
			s.pubsub.Pub(events...)
		}
		return nil
	})
}

// Do function on the context of the transaction.
// The transaction begins before run the txFunc, then run the txFunc, then commit if txFunc returns nil.
// The transaction is rollbacked if txFunc returns some error.
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

// It creates room specified by given actiom message.
// It returns created Room's ID and error if any.
func (s *CommandServiceImpl) CreateRoom(ctx context.Context, m action.CreateRoom) (roomID uint64, err error) {
	user, err := s.users.Find(ctx, m.SenderID)
	if err != nil {
		return 0, err
	}

	err = s.withEventTransaction(ctx, s.rooms, func(ctx context.Context) ([]event.Event, error) {
		room, err := domain.NewRoom(
			ctx, s.rooms, m.RoomName,
			&user, domain.NewUserIDSet(m.RoomMemberIDs...),
		)
		if err != nil {
			return nil, err
		}
		roomID = room.ID

		return room.Events(), nil
	})
	return roomID, err
}

// It deletes room specified by given actiom message. It returns deleted Room's ID and
// error if any.
func (s *CommandServiceImpl) DeleteRoom(ctx context.Context, m action.DeleteRoom) (roomID uint64, err error) {
	user, err := s.users.Find(ctx, m.SenderID)
	if err != nil {
		return 0, err
	}

	err = s.withEventTransaction(ctx, s.rooms, func(ctx context.Context) ([]event.Event, error) {
		room, err := s.rooms.Find(ctx, m.RoomID)
		if err != nil {
			return nil, err
		}
		// room ID to delete
		roomID = room.ID

		err = room.Delete(ctx, s.rooms, &user)
		if err != nil {
			return nil, err
		}

		return room.Events(), nil
	})
	return roomID, err
}

// Post the message to the specified room.
// It returns posted message id and nil or error
// which indicates the message can not be posted.
func (s *CommandServiceImpl) PostRoomMessage(ctx context.Context, m action.ChatMessage) (msgID uint64, err error) {
	room, err := s.rooms.Find(ctx, m.RoomID)
	if err != nil {
		return 0, err
	}

	user, err := s.users.Find(ctx, m.SenderID)
	if err != nil {
		return 0, err
	}

	err = s.withEventTransaction(ctx, s.msgs, func(ctx context.Context) ([]event.Event, error) {
		msg, err := domain.NewRoomMessage(ctx, s.msgs, user, room, m.Content)
		if err != nil {
			return nil, err
		}
		msgID = msg.ID

		return msg.Events(), nil
	})
	return msgID, err
}

// Mark the message is read by the specified user.
// It returns updated room ID or error when the message can not be marked to read.
func (s *CommandServiceImpl) ReadRoomMessages(ctx context.Context, m action.ReadMessages) (uint64, error) {
	// set default value insteadly.
	if m.ReadAt.Equal(time.Time{}) {
		m.ReadAt = time.Now()
	}

	user, err := s.users.Find(ctx, m.SenderID)
	if err != nil {
		return 0, err
	}

	txErr := s.withEventTransaction(ctx, s.rooms, func(ctx context.Context) ([]event.Event, error) {
		room, err := s.rooms.Find(ctx, m.RoomID)
		if err != nil {
			return nil, err
		}

		if _, err := room.ReadMessagesBy(&user, m.ReadAt); err != nil {
			return nil, err
		}
		if _, err := s.rooms.Store(ctx, room); err != nil {
			return nil, err
		}
		return room.Events(), nil
	})
	return m.RoomID, txErr
}
