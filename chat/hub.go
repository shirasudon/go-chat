package chat

import (
	"context"
	"fmt"
	"log"

	"github.com/shirasudon/go-chat/chat/action"
	"github.com/shirasudon/go-chat/domain"
	"github.com/shirasudon/go-chat/domain/event"
)

// Hub accepts any user connections to
// propagates domain events for those connections.
type Hub struct {
	messages chan actionMessageRequest
	events   chan event.Event
	shutdown chan struct{}

	chatCommand   *CommandService
	chatQuery     *QueryService
	activeClients *domain.ActiveClientRepository
	pubsub        Pubsub
}

// actionMessageRequest is a composit struct of
// ActionMessage and Conn to send the message.
// It is used to handle ActionMessage by ChatHub.
type actionMessageRequest struct {
	action.ActionMessage
	Conn domain.Conn
}

func NewHub(cmdService *CommandService, queryService *QueryService) *Hub {
	if cmdService == nil || queryService == nil {
		panic("passed either nil services")
	}

	return &Hub{
		messages: make(chan actionMessageRequest, 1),
		events:   make(chan event.Event, 1),
		shutdown: make(chan struct{}),

		chatCommand:   cmdService,
		chatQuery:     queryService,
		activeClients: domain.NewActiveClientRepository(64),
		pubsub:        cmdService.pubsub,
	}
}

// Stop handling messages from the connections and
// sending events to connections.
// Multiple calling will cause panic.
func (hub *Hub) Shutdown() {
	close(hub.shutdown)
}

// Start handling messages from the connections and
// sending events to connections.
// It will blocks untill
// called Shutdown() or context is done.
func (hub *Hub) Listen(ctx context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// run service for sending event to connections.
	go hub.eventSendingService(ctx)

	for {
		select {
		case req := <-hub.messages:
			err := hub.handleMessage(ctx, req)
			if err != nil {
				log.Println(err)
				// TODO send error event to requested connection.
				// req.Conn.Send()
			}

		case <-hub.shutdown:
			return
		case <-ctx.Done():
			return
		}
	}
}

func (hub *Hub) broadcastEvent(ev event.Event, targetIDs ...uint64) error {
	if len(targetIDs) == 0 {
		return nil
	}

	acs, err := hub.activeClients.FindAllByUserIDs(targetIDs)
	if err != nil {
		return err
	}

	toSend := NewEventJSON(ev)
	for _, ac := range acs {
		ac.Send(toSend)
	}
	return nil
}

func (hub *Hub) eventSendingService(ctx context.Context) {
	events := hub.pubsub.Sub(
		event.TypeMessageCreated,
		event.TypeActiveClientActivated,
		event.TypeActiveClientInactivated,
	)

	for {
		select {
		case <-hub.shutdown:
			return
		case <-ctx.Done():
			return
		case ev, chAlived := <-events:
			if !chAlived {
				return
			}

			switch ev := ev.(type) {
			case event.MessageCreated:
				// send activate event for all friends.
				room, err := hub.chatQuery.rooms.Find(ctx, ev.RoomID)
				if err != nil {
					// TODO error handling
					log.Println(err)
					continue
				}

				err = hub.broadcastEvent(ev, room.MemberIDs()...)
				if err != nil {
					// TODO error handling
					log.Println(err)
					continue
				}

			case event.ActiveClientActivated:
				// send activate event for all friends.
				user, err := hub.chatQuery.users.Find(ctx, ev.UserID)
				if err != nil {
					// TODO error handling
					log.Println(err)
					continue
				}

				targetIDs := append(user.FriendIDs.List(), user.ID) // contains user-self.
				err = hub.broadcastEvent(ev, targetIDs...)
				if err != nil {
					// TODO error handling
					log.Println(err)
					continue
				}

			case event.ActiveClientInactivated:
				// send inactivate event for all friends.
				user, err := hub.chatQuery.users.Find(ctx, ev.UserID)
				if err != nil {
					// TODO error handling
					log.Println(err)
					continue
				}

				err = hub.broadcastEvent(ev, user.FriendIDs.List()...)
				if err != nil {
					// TODO error handling
					log.Println(err)
					continue
				}
			}
		}
	} // ... for
}

func (hub *Hub) handleMessage(ctx context.Context, req actionMessageRequest) error {
	var err error = nil

	if !hub.activeClients.ExistByConn(req.Conn) {
		return fmt.Errorf("not connected to the server")
	}

	switch m := req.ActionMessage.(type) {
	case action.ChatMessage:
		_, err = hub.chatCommand.PostRoomMessage(ctx, m)
	// TODO case action.EditChatMessage:
	// TODO case action.DeleteChatMessage:
	case action.ReadMessage:
		err = hub.chatCommand.ReadRoomMessage(ctx, m)
	case action.TypeStart, action.TypeEnd:
		// TODO convert acitionMessage to event then publish in chatCommand
	}

	return err
}

// Send ActionMessage with the connection which sent the message.
// the connection is used to verify that the message is exactlly
// sent by the connected user.
// The error is sent to given conn when the message is invalid.
func (hub *Hub) Send(conn domain.Conn, message action.ActionMessage) {
	select {
	case <-hub.shutdown:
		return
	case hub.messages <- actionMessageRequest{message, conn}:
	}
}

// Connect new websocket connection to the hub.
func (hub *Hub) Connect(ctx context.Context, c domain.Conn) error {
	userID := c.UserID()
	user, err := hub.chatQuery.users.Find(ctx, userID)
	if err != nil {
		return fmt.Errorf("connected user(%d) is not found", userID)
	}

	if ac, err := hub.activeClients.Find(userID); err == nil {
		// the connected user is already activated, so just add connection.
		_, err := ac.AddConn(c)
		if err != nil {
			return err
		}
		return hub.activeClients.Store(ac)
	}

	// the conncected user is inactivate, activate it.
	_, activated, err := domain.NewActiveClient(hub.activeClients, c, user)
	if err != nil {
		return err
	}

	// publish activated event.
	hub.pubsub.Pub(activated)
	return nil
}

// Disconnect the given websocket connection from the hub.
// it will no-operation when non-connected connection is given.
func (hub *Hub) Disconnect(conn domain.Conn) {
	ac, err := hub.activeClients.Find(conn.UserID())
	if err != nil {
		// not connected, no operations
		return
	}

	connN, err := ac.RemoveConn(conn)
	if err != nil {
		// conn is not found in active client, no operation
		return
	}

	if connN > 0 {
		// connection still exist in active client, no operation
		return
	}

	// any connection does not exist in active client
	// remove AcitiveClient from the repository.
	inactivated, err := ac.Delete(hub.activeClients)
	if err != nil {
		// TODO error log
		return
	}

	// publish inactivated event.
	hub.pubsub.Pub(inactivated)
}
