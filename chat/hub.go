package chat

import (
	"context"
	"fmt"
	"log"

	"github.com/shirasudon/go-chat/chat/action"
	"github.com/shirasudon/go-chat/domain"
	"github.com/shirasudon/go-chat/domain/event"
)

// Conn is exported at the chat package so that the higher layer need not to import domain package.
type Conn = domain.Conn

// Hub is a interface for a hub which conmmunicates the event/action messages
// between the active client connections.
type Hub interface {
	// Connect accepts new connection to the hub.
	// If conection is invalid then return error.
	Connect(ctx context.Context, c Conn) error

	// Send sends ActionMessage with the connection which sent the message, to the hub.
	// The Conn is used to verify that the message is exactlly
	// sent by the connected user.
	// The error is sent to given conn when the message is invalid.
	Send(conn Conn, message action.ActionMessage)

	// Disconnect disconnects the given connection from the hub.
	// it will no-operation when non-connected connection is given.
	Disconnect(conn Conn)
}

// HubImpl accepts any user connections to
// propagates domain events for those connections.
// It implements Hub interface.
type HubImpl struct {
	messages chan actionMessageRequest
	events   chan event.Event
	shutdown chan struct{}

	chatCommand   *CommandServiceImpl
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

func NewHubImpl(cmd *CommandServiceImpl) *HubImpl {
	if cmd == nil {
		panic("passed nil arguments")
	}

	return &HubImpl{
		messages: make(chan actionMessageRequest, 1),
		events:   make(chan event.Event, 1),
		shutdown: make(chan struct{}),

		chatCommand:   cmd,
		activeClients: domain.NewActiveClientRepository(64),
		pubsub:        cmd.pubsub,
	}
}

// Stop handling messages from the connections and
// sending events to connections.
// Multiple calling will cause panic.
func (hub *HubImpl) Shutdown() {
	close(hub.shutdown)
}

// Start handling messages from the connections and
// sending events to connections.
// It will blocks untill
// called Shutdown() or context is done.
func (hub *HubImpl) Listen(ctx context.Context) {
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

func (hub *HubImpl) broadcastEvent(ev event.Event, targetIDs ...uint64) error {
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

// HubHandlingEventTypes a list of event types to be handled
// by the Hub interface.
var HubHandlingEventTypes = []event.Type{
	event.TypeMessageCreated,
	event.TypeActiveClientActivated,
	event.TypeActiveClientInactivated,
	event.TypeRoomCreated,
	event.TypeRoomDeleted,
	event.TypeRoomAddedMember,
	event.TypeRoomRemovedMember,
	event.TypeRoomMessagesReadByUser,
}

func (hub *HubImpl) eventSendingService(ctx context.Context) {
	events := hub.pubsub.Sub(HubHandlingEventTypes...)

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
			if ev, ok := ev.(event.Event); ok {
				err := hub.sendEvent(ctx, ev)
				if err != nil {
					// TODO error handling
					log.Println(err)
				}
			}
		}
	} // ... for
}

func (hub *HubImpl) sendEvent(ctx context.Context, ev event.Event) error {
	var (
		chatCommand = hub.chatCommand
		targetIDs   = []uint64{}
	)

	// TODO: integrates these into interfaces, RoomMemberSender and UserFriendSender?
	switch ev := ev.(type) {
	case event.MessageCreated:
		room, err := chatCommand.rooms.Find(ctx, ev.RoomID)
		if err != nil {
			return err
		}
		targetIDs = room.MemberIDSet.List()

	case event.RoomCreated:
		room, err := chatCommand.rooms.Find(ctx, ev.RoomID)
		if err != nil {
			return err
		}
		targetIDs = room.MemberIDSet.List()

	case event.RoomDeleted:
		room, err := chatCommand.rooms.Find(ctx, ev.RoomID)
		if err != nil {
			return err
		}
		targetIDs = room.MemberIDSet.List()

	case event.RoomAddedMember:
		room, err := chatCommand.rooms.Find(ctx, ev.RoomID)
		if err != nil {
			return err
		}
		targetIDs = room.MemberIDSet.List()

	case event.RoomRemovedMember:
		room, err := chatCommand.rooms.Find(ctx, ev.RoomID)
		if err != nil {
			return err
		}
		targetIDs = room.MemberIDSet.List()

	case event.RoomMessagesReadByUser:
		room, err := chatCommand.rooms.Find(ctx, ev.RoomID)
		if err != nil {
			return err
		}
		targetIDs = room.MemberIDSet.List()

	case event.ActiveClientActivated:
		user, err := chatCommand.users.Find(ctx, ev.UserID)
		if err != nil {
			return err
		}
		targetIDs = append(user.FriendIDs.List(), user.ID) // contains user-self.

	case event.ActiveClientInactivated:
		user, err := chatCommand.users.Find(ctx, ev.UserID)
		if err != nil {
			return err
		}
		targetIDs = user.FriendIDs.List()
	}

	return hub.broadcastEvent(ev, targetIDs...)
}

func (hub *HubImpl) handleMessage(ctx context.Context, req actionMessageRequest) error {
	var err error = nil

	if !hub.activeClients.ExistByConn(req.Conn) {
		return fmt.Errorf("not connected to the server")
	}

	switch m := req.ActionMessage.(type) {
	case action.ChatMessage:
		_, err = hub.chatCommand.PostRoomMessage(ctx, m)
	// TODO case action.EditChatMessage:
	// TODO case action.DeleteChatMessage:
	case action.ReadMessages:
		_, err = hub.chatCommand.ReadRoomMessages(ctx, m)
	case action.TypeStart, action.TypeEnd:
		// TODO convert acitionMessage to event then publish in chatCommand
	}

	return err
}

// Send ActionMessage with the connection which sent the message.
// the connection is used to verify that the message is exactlly
// sent by the connected user.
// The error is sent to given conn when the message is invalid.
func (hub *HubImpl) Send(conn Conn, message action.ActionMessage) {
	select {
	case <-hub.shutdown:
		return
	case hub.messages <- actionMessageRequest{message, conn}:
	}
}

// Connect new websocket connection to the hub.
func (hub *HubImpl) Connect(ctx context.Context, c Conn) error {
	userID := c.UserID()
	user, err := hub.chatCommand.users.Find(ctx, userID)
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
func (hub *HubImpl) Disconnect(conn Conn) {
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
