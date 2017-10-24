package chat

import (
	"context"
	"log"

	"github.com/shirasudon/go-chat/chat/action"
	"github.com/shirasudon/go-chat/domain"
)

// Hub is the hub which accepts any websocket connections to
// serve chat messages for the other connections.
// any websocket connections connect this hub firstly, then the
// connections are managed by Hub.
type Hub struct {
	connects    chan Conn
	disconnects chan Conn
	messages    chan actionMessageRequest
	errors      chan error

	chatCommand    *CommandService
	chatQuery      *QueryService
	messageHandler *messageHandler
}

// actionMessageRequest is a composit struct of
// ActionMessage and Conn to send the message.
// It is used to handle ActionMessage by ChatHub.
type actionMessageRequest struct {
	action.ActionMessage
	Conn Conn
}

func NewHub(repos domain.Repositories, pubsub Pubsub) *Hub {
	chatCommand := NewCommandService(repos, pubsub)
	chatQuery := NewQueryService(repos)
	return &Hub{
		connects:       make(chan Conn, 1),
		disconnects:    make(chan Conn, 1),
		messages:       make(chan actionMessageRequest, 1),
		errors:         make(chan error, 1),
		chatCommand:    chatCommand,
		chatQuery:      chatQuery,
		messageHandler: newMessageHandler(chatCommand, chatQuery),
	}
}

func (hub *Hub) Listen(ctx context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	for {
		// disconnect event has priority than others
		// because disconnected client can not be received any message.
		select {
		case c := <-hub.disconnects:
			if err := hub.disconnectClient(ctx, c); err != nil {
				// TODO err handling
				log.Printf("Disonnect Error: %v\n", err)
			}
			continue
		case <-ctx.Done():
			return
		default:
			// fall through if no disconnect events.
		}

		select {
		case c := <-hub.connects:
			if err := hub.connectClient(ctx, c); err != nil {
				// TODO: error handling
				log.Printf("Connect Error: %v\n", err)
			}

		case c := <-hub.disconnects:
			if err := hub.disconnectClient(ctx, c); err != nil {
				// TODO err handling
				log.Printf("Disonnect Error: %v\n", err)
			}

		case req := <-hub.messages:
			if err := hub.messageHandler.handleMessage(ctx, req); err != nil {
				// TODO err handling
				sendError(req.Conn, err, req.ActionMessage)
				log.Printf("Message Error: %v\n", err)
			}

		case err := <-hub.errors:
			// TODO error handling
			log.Printf("Error: %v\n", err)

		case <-ctx.Done():
			return
		}
	}
}

// Connect new websocket connection to the hub.
func (hub *Hub) Connect(conn Conn) {
	hub.connects <- conn
}

// Disconnect the given websocket connection from the hub.
// it will no-operation when non-connected connection is given.
func (hub *Hub) Disconnect(conn Conn) {
	hub.connects <- conn
}

// Send ActionMessage with the connection which sent the message.
// the connection is used to verify that the message is exactlly
// sent by the connected connection.
func (hub *Hub) Send(conn Conn, message action.ActionMessage) {
	hub.messages <- actionMessageRequest{message, conn}
}

func (hub *Hub) connectClient(ctx context.Context, c Conn) error {
	return hub.messageHandler.connectClient(ctx, c)
}

func (hub *Hub) disconnectClient(ctx context.Context, c Conn) error {
	return hub.messageHandler.disconnectClient(ctx, c)
}

func sendError(c Conn, err error, cause ...action.ActionMessage) {
	log.Println(err)
	go func() { c.Send(action.NewErrorMessage(err, cause...)) }()
}
