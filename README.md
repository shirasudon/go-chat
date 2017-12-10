# chat

[![CircleCI](https://circleci.com/gh/shirasudon/go-chat/tree/master.svg?style=svg)](https://circleci.com/gh/shirasudon/go-chat/tree/master)
[![codecov](https://codecov.io/gh/shirasudon/go-chat/branch/master/graph/badge.svg)](https://codecov.io/gh/shirasudon/go-chat)

Backend chat server based on the Websocket written by Go.

## Websocket Connection

The server can accepts the Websocket connetion at `/chat/ws`.
The Websocket connetion can be used two ways:

1. Receive events from the server.
1. Send actions to the server.

### Receive events

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

### Send actions

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

## REST API

* GetUserInfo -- `GET /chat/users/:user_id`

It returns user information specified by `user_id`.

Request JSON data: `None`.

Responce JSON:

```javascript
{
    "user_id": user_id,
    "user_name": "<user name>",
    "first_name": "<first name>",
    "last_name": "<last name>",

    "friends": [
        {
            "user_id": user_id,
            "user_name": "<user name>",
            "first_name": "<first name>",
            "last_name": "<last name>",
        },
        {
            ...
        }
    ],

    "rooms": [
        {
            "room_id": room_id,
            "room_name": "<room name>",
        },
        {
            ...
        }
    ]
}
```

---

* GetRoomInfo -- `GET /chat/rooms/:room_id`

It returns room information specified by `room_id`.

Request JSON data: `None`.

Responce JSON:

```javascript
{
    "room_id": room_id,
    "room_name": "<room name>",
    "room_creator_id": room_creator_id, // user_id

    "room_members": [
        {
            "user_id": user_id,
            "user_name": "<user name>",
            "first_name": "<first name>",
            "last_name": "<last name>",
        },
        {
            ...
        }
    ],

    "room_members_size": room_members_size,
}
```

---

* GetUnreadRoomMessage -- `GET /chat/rooms/:room_id/messages/unread`

It returns messages unread by the logged-in user in the room specified by `room_id`.

Request JSON: 

```javascript
{
    "limit": limit_number,
}
```


Responce JSON:

```javascript
{
    "room_id": room_id,

    "messages": [
        {
            "message_id": message_id,
            "content":    "<message content>",
            "created_at": created_at,
        },
        ...
    ],

    "messages_size": messages_size,
}
```

