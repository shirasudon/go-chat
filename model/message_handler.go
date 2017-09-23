package model

import (
	"context"

	"github.com/shirasudon/go-chat/entity"
	"github.com/shirasudon/go-chat/model/action"
)

type messageHandler struct {
	rooms   *RoomManager
	clients *ClientManager

	msgRepo entity.MessageRepository
}

func newMessageHandler(repos entity.Repositories) *messageHandler {
	return &messageHandler{
		rooms:   NewRoomManager(repos),
		clients: NewClientManager(repos),
		msgRepo: repos.Messages(),
	}
}

func (handler *messageHandler) connectClient(ctx context.Context, c Conn) error {
	if err := handler.clients.connectClient(ctx, c); err != nil {
		return err
	}
	if err := handler.rooms.connectClient(ctx, c.UserID()); err != nil {
		return err
	}
	return nil
}

func (handler *messageHandler) disconnectClient(ctx context.Context, c Conn) error {
	handler.clients.disconnectClient(c)
	if err := handler.rooms.disconnectClient(ctx, c.UserID()); err != nil {
		return err
	}
	return nil
}

func (handler *messageHandler) broadcastsRoomMembers(roomID uint64, m action.ActionMessage) {
	memberIDs := handler.rooms.roomMemberIDs(roomID)
	handler.clients.broadcastsUsers(memberIDs, m)
}

func (handler *messageHandler) broadcastsFriends(userID uint64, m action.ActionMessage) {
	activeC, ok := handler.clients.clients[userID]
	if ok {
		handler.clients.broadcastsFriends(activeC, m)
	}
}

func (handler *messageHandler) handleMessage(ctx context.Context, req actionMessageRequest) error {
	// TODO set UserID to req.ActionMessage
	switch m := req.ActionMessage.(type) {
	case action.ChatActionMessage:
		if err := handler.clients.validateClientHasRoom(req.Conn, m.GetSenderID(), m.GetRoomID()); err != nil {
			return err
		}
		return handler.handleChatActionMessage(ctx, req.Conn, m)
	}

	return nil
}

func (handler *messageHandler) handleChatActionMessage(ctx context.Context, conn Conn, m action.ChatActionMessage) error {
	switch m := m.(type) {
	case action.ChatMessage:
		var err error
		m.ID, err = handler.msgRepo.Add(ctx, entity.Message{
			Content: m.Content,
			UserID:  m.SenderID,
			RoomID:  m.RoomID,
		})
		if err != nil {
			return err
		}
		handler.broadcastsRoomMembers(m.RoomID, m)

	case action.ReadMessage:
		if err := handler.msgRepo.ReadMessage(ctx, m.RoomID, m.SenderID, m.MessageIDs); err != nil {
			return err
		}

		handler.broadcastsRoomMembers(m.RoomID, m)

	case action.TypeStart, action.TypeEnd:
		handler.broadcastsRoomMembers(m.GetRoomID(), m)
	}
	return nil
}
