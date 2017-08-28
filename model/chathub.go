package model

import (
	"context"

	"github.com/shirasudon/go-chat/entity"
)

type ChatHub struct {
	rooms   *RoomManager
	clients *ClientManager

	msgRepo entity.MessageRepository
}

func NewChatHub(repos entity.Repositories) *ChatHub {
	return &ChatHub{
		rooms:   NewRoomManager(repos),
		clients: NewClientManager(repos),
		msgRepo: repos.Messages(),
	}
}

func (hub *ChatHub) connectClient(ctx context.Context, c *Conn) error {
	if err := hub.clients.connectClient(ctx, c); err != nil {
		return err
	}
	if err := hub.rooms.connectClient(ctx, c.userID); err != nil {
		return err
	}
	return nil
}

func (hub *ChatHub) disconnectClient(ctx context.Context, c *Conn) error {
	hub.clients.disconnectClient(c)
	if err := hub.rooms.disconnectClient(ctx, c.userID); err != nil {
		return err
	}
	return nil
}

func (hub *ChatHub) broadcastsRoomMembers(roomID uint64, m ActionMessage) {
	memberIDs := hub.rooms.roomMemberIDs(roomID)
	hub.clients.broadcastsUsers(memberIDs, m)
}

func (hub *ChatHub) broadcastsFriends(userID uint64, m ActionMessage) {
	activeC, ok := hub.clients.clients[userID]
	if ok {
		hub.clients.broadcastsFriends(activeC, m)
	}
}

func (hub *ChatHub) handleMessage(ctx context.Context, req actionMessageRequest) error {
	switch m := req.ActionMessage.(type) {
	case ChatActionMessage:
		if err := hub.clients.validateClientHasRoom(req.Conn, m.GetSenderID(), m.GetRoomID()); err != nil {
			return err
		}
		return hub.handleChatActionMessage(ctx, req.Conn, m)
	}

	return nil
}

func (hub *ChatHub) handleChatActionMessage(ctx context.Context, conn *Conn, m ChatActionMessage) error {
	switch m := m.(type) {
	case ChatMessage:
		var err error
		m.ID, err = hub.msgRepo.Add(ctx, entity.Message{
			Content: m.Content,
			UserID:  m.SenderID,
			RoomID:  m.RoomID,
		})
		if err != nil {
			return err
		}
		hub.broadcastsRoomMembers(m.RoomID, m)

	case ReadMessage:
		if err := hub.msgRepo.ReadMessage(ctx, m.RoomID, m.SenderID, m.MessageIDs); err != nil {
			return err
		}

		hub.broadcastsRoomMembers(m.RoomID, m)

	case TypeStart, TypeEnd:
		hub.broadcastsRoomMembers(m.GetRoomID(), m)
	}
	return nil
}
