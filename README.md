# go-chat

[![CircleCI](https://circleci.com/gh/shirasudon/go-chat/tree/master.svg?style=svg)](https://circleci.com/gh/shirasudon/go-chat/tree/master)
[![codecov](https://codecov.io/gh/shirasudon/go-chat/branch/master/graph/badge.svg)](https://codecov.io/gh/shirasudon/go-chat)

Backend chat server based on the Websocket written by Go.

## How to install

Installig go-chat requires Go 1.9 or newer, and [dep](https://github.com/golang/dep) command.

You can download `go-chat` with `go` and `dep` commands:

```bash
$ go get -u github.com/shirasudon/go-chat
$ cd $GOPATH/src/github.com/shirasudon/go-chat
$ dep ensure
```

To start stand-alone local server, run the followings:

```bash
$ go run main/main.go
```

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

### CreateRoom -- `POST /chat/rooms`

It creates new chat room.

Request JSON: 

```javascript
{
    "room_name": "<room_name>",
    "room_member_ids": [1,2, ...],
}
```

response JSON:

```javascript
{
    "room_id": created_room_id,
    "ok": true,
}
```

### DeleteRoom -- `DELETE /chat/rooms/:room_id`

It deletes existance chat room specified by `room_id`.

Request JSON: `None`.

response JSON:

```javascript
{
    "room_id": deleted_room_id,
    "ok": true,
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
            "message_read_at", message_read_time,
        },
        {
            ...
        }
    ],

    "room_members_size": room_members_size,
}
```

### GetRoomMessage -- `GET /chat/rooms/:room_id/messages`

It returns messages in the room specified by `room_id`.
The returned messages contains both read and unread messages.

Query Paramters:

* `before`: The start point of the `created_at` in result messages. It must be RFC3339 form.
* `limit`: The number of result messages.

Request JSON: 

```javascript
{
    "before": "time with RFC3339 format",
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

Example:

`GET /chat/rooms/:room_id/messages?before=2018-01-01T12:34:56Z?limit=10` will returns 
queried result which contains 10 messages and all of these are created before 2018/01/01 12:34:56.


### GetUnreadRoomMessage -- `GET /chat/rooms/:room_id/messages/unread`

It returns messages unread by the logged-in user in the room specified by `room_id`.

Query Paramters:

* `limit`: The number of result messages.

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

Example:

`GET /chat/rooms/:room_id/messages/unread?limit=10` will returns queried result which contains 10 messages.


### ReadRoomMessages -- `POST /chat/rooms/:room_id/messages/read`

It notifies to the server that the messages in the room specified by the `room_id` are
read by the user.

Request JSON: 

```javascript
{
    "read_at": <messages read time> // time format
}
```


response JSON:

```javascript
{
    "updated_room_id": room_id,
    "read_user_id": user_id,
    "ok": true or false,
}
```

