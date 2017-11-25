package chat

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/labstack/echo"

	"github.com/shirasudon/go-chat/chat"
	"github.com/shirasudon/go-chat/chat/action"
)

var (
	// ErrRequireLoginFirst indicates that unauthenticated error message.
	ErrRequireLoginFirst = errors.New("require login first")

	// ErrAPIRequireLoginFirst indicates that unauthenticated error message with htto status code.
	ErrAPIRequireLoginFirst = NewHTTPError(http.StatusForbidden, ErrRequireLoginFirst.Error())
)

type RESTHandler struct {
	chatCmd   *chat.CommandService
	chatQuery *chat.QueryService
}

func NewRESTHandler(chatCmd *chat.CommandService, chatQuery *chat.QueryService) *RESTHandler {
	return &RESTHandler{
		chatCmd:   chatCmd,
		chatQuery: chatQuery,
	}
}

// parse uint parameter with key from echo.Context.
// it returns parsed uint number, or 0 if can not parsed.
func uintParam(e echo.Context, key string) uint64 {
	n, err := strconv.ParseUint(e.Param(key), 10, 64)
	if err != nil {
		return 0
	}
	return n
}

func (rest *RESTHandler) validateUserID(e echo.Context) (uint64, error) {
	userID, ok := LoggedInUserID(e)
	if !ok {
		return 0, ErrAPIRequireLoginFirst
	}

	userPathID, err := strconv.ParseUint(e.Param("id"), 10, 64)
	if err != nil || userPathID != userID {
		return 0, NewHTTPError(http.StatusBadRequest, fmt.Errorf("requested user id(%v) is not allowed", userPathID))
	}

	return userID, nil
}

func (rest *RESTHandler) validateParamRoomID(e echo.Context) (uint64, error) {
	const ParamKey = "room_id"

	param := e.Param(ParamKey)
	roomID, err := strconv.ParseUint(param, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("requested room id(%v) is not allowed", param)
	}

	return roomID, nil
}

func (rest *RESTHandler) CreateRoom(e echo.Context) error {
	userID, err := rest.validateUserID(e)
	if err != nil {
		return err
	}

	createRoom := action.CreateRoom{}
	if err = e.Bind(&createRoom); err != nil {
		// TODO return error as JSON format
		return NewHTTPError(http.StatusBadRequest, err)
	}
	createRoom.SenderID = userID

	createdID, err := rest.chatCmd.CreateRoom(e.Request().Context(), createRoom)
	if err != nil {
		return NewHTTPError(http.StatusInternalServerError, err)
	}

	response := struct {
		RoomID uint64 `json:"room_id"`
		OK     bool   `json:"ok"`
	}{
		RoomID: createdID,
		OK:     true,
	}
	return e.JSON(http.StatusCreated, response)
}

func (rest *RESTHandler) DeleteRoom(e echo.Context) error {
	userID, err := rest.validateUserID(e)
	if err != nil {
		return err
	}

	deleteRoom := action.DeleteRoom{}
	if err := e.Bind(&deleteRoom); err != nil {
		// TODO return error as JSON format
		return NewHTTPError(http.StatusBadRequest, err)
	}
	deleteRoom.SenderID = userID

	deletedID, err := rest.chatCmd.DeleteRoom(e.Request().Context(), deleteRoom)
	if err != nil {
		return NewHTTPError(http.StatusInternalServerError, err)
	}

	response := struct {
		RoomID uint64 `json:"room_id"`
		OK     bool   `json:"ok"`
	}{
		RoomID: deletedID,
		OK:     true,
	}
	return e.JSON(http.StatusNoContent, response)
}

func (rest *RESTHandler) GetRoomInfo(e echo.Context) error {
	userID, ok := LoggedInUserID(e)
	if !ok {
		return ErrAPIRequireLoginFirst
	}

	roomID, err := rest.validateParamRoomID(e)
	if err != nil {
		return NewHTTPError(http.StatusBadRequest, err)
	}

	info, err := rest.chatQuery.FindRoomInfo(e.Request().Context(), userID, roomID)
	if err != nil {
		// TODO distinguish logic error or infra error
		// if _, ok := err.(*chat.InfraError); ok {
		return NewHTTPError(http.StatusInternalServerError, err)
	}

	return e.JSON(http.StatusFound, info)
}

func (rest *RESTHandler) GetUserInfo(e echo.Context) error {
	userID := uintParam(e, "user_id")
	relation, err := rest.chatQuery.FindUserRelation(e.Request().Context(), userID)
	if err != nil {
		return err
	}

	return e.JSON(http.StatusFound, relation)
}

func (rest *RESTHandler) PostRoomMessage(e echo.Context) error {
	userID, ok := LoggedInUserID(e)
	if !ok {
		return ErrAPIRequireLoginFirst
	}

	postMsg := action.ChatMessage{}
	if err := e.Bind(&postMsg); err != nil {
		// TODO return error as JSON format
		return NewHTTPError(http.StatusBadRequest, err)
	}
	postMsg.SenderID = userID
	postMsg.RoomID = uintParam(e, "room_id") // TODO use const variable for "room_id"
	msgID, err := rest.chatCmd.PostRoomMessage(e.Request().Context(), postMsg)
	if err != nil {
		return NewHTTPError(http.StatusInternalServerError, err)
	}

	response := struct {
		MsgID  uint64 `json:"message_id"`
		RoomID uint64 `json:"room_id"`
		OK     bool   `json:"ok"`
	}{
		MsgID:  msgID,
		RoomID: postMsg.RoomID,
		OK:     true,
	}
	return e.JSON(http.StatusCreated, response)
}

func (rest *RESTHandler) GetRoomMessages(e echo.Context) error {
	userID, ok := LoggedInUserID(e)
	if !ok {
		return ErrAPIRequireLoginFirst
	}

	qRoomMsg := action.QueryRoomMessages{}
	if err := e.Bind(&qRoomMsg); err != nil {
		// TODO return error as JSON format
		return NewHTTPError(http.StatusBadRequest, err)
	}

	roomMsg, err := rest.chatQuery.FindRoomMessages(e.Request().Context(), userID, qRoomMsg)
	if err != nil {
		return NewHTTPError(http.StatusInternalServerError, err)
	}

	return e.JSON(http.StatusFound, roomMsg)
}
