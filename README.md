# chat

[![CircleCI](https://circleci.com/gh/shirasudon/go-chat/tree/master.svg?style=svg)](https://circleci.com/gh/shirasudon/go-chat/tree/master)

Backend chat server based on the Websocket written by Go.

# Websocket Connection

The server can accepts the Websocket connetion at `/chat/ws`.
The Websocket connetion can be used two ways:

1. Receive events from the server.
1. Send actions to the server.

## Receive events

The Websocket connetion can be used as the Event stream which 
publish the Events, such as Message created, Message deleted and so on, 
to the client. The client can subscribe such events.

The event format is a JSON like:

```javascript
{
  "event": "<event name>",
  "data": {
    xxx,
    yyy
  }
}
```

## Send actions

The Websocket connetion can be used as the chat application interface
for sending the commands, such as Post new message, Create new room and so on,
to the server. 

The action format is a JSON like:

```javascript
{
  "action": "<action name>",
  "data": {
    xxx,
    yyy
  }
}
```

Note that the responses to those commands are indirectly returnd by the events.


