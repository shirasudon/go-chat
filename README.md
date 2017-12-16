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

### Login -- `POST /login`

It login to the chat application.
The login session is stored to the cookie.

User should login first and use cookie to access chat API.


Request JSON: 

```javascript
{
    "name": "user name",
    "password": "password",
    "remember_me": true or false,
}
```

Response JSON:

```javascript
{
    "logged_in": true or false,
    "remember_me": true or false,
    "user_id": <logged-in user ID>, // number
    "error": "error message if any",
}
```

### GetLoginState `GET /login`

It gets current login state.

Request JSON: `None`

Response JSON:

```javascript
{
    "logged_in": true or false,
    "remember_me": true or false,
    "user_id": <logged-in user ID>, // number
    "error": "error message if any",
}
```

### Logout `POST /logout`

It logout from the chat application.

Request JSON: `None`

Response JSON:

```javascript
{
    "logged_in": false, 
    "error": "error message if any",
}
```

### GetUserInfo -- `GET /chat/users/:user_id`

It returns user information specified by `user_id`.

Request JSON: `None`.

response JSON:

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

### GetRoomInfo -- `GET /chat/rooms/:room_id`

It returns room information specified by `room_id`.

Request JSON: `None`.

response JSON:

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

### GetUnreadRoomMessage -- `GET /chat/rooms/:room_id/messages/unread`

It returns messages unread by the logged-in user in the room specified by `room_id`.

Request JSON: 

```javascript
{
    "limit": limit_number,
}
```


response JSON:

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

