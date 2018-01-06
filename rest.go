package chat

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/labstack/echo"

	"github.com/shirasudon/go-chat/chat"
	"github.com/shirasudon/go-chat/chat/action"
)

var (
	// ErrAPIRequireLoginFirst indicates that unauthenticated error message with htto status code.
	ErrAPIRequireLoginFirst = NewHTTPError(http.StatusForbidden, "require login first")
)

type RESTHandler struct {
	chatCmd   chat.CommandService
	chatQuery chat.QueryService
}

func NewRESTHandler(chatCmd chat.CommandService, chatQuery chat.QueryService) *RESTHandler {
	return &RESTHandler{
		chatCmd:   chatCmd,
		chatQuery: chatQuery,
	}
}

const (
	// keys for the URL parameters. e.g. /root/:param_name
	ParamKeyUserID = "user_id"
	ParamKeyRoomID = "room_id"
)

func validateParamUserID(e echo.Context) (uint64, error) {
	param := e.Param(ParamKeyUserID)
	userID, err := strconv.ParseUint(param, 10, 64)
	if err != nil {
		return 0, NewHTTPError(http.StatusBadRequest, fmt.Errorf("requested user id(%v) is not allowed", param))
	}

	return userID, nil
}

func validateParamRoomID(e echo.Context) (uint64, error) {
	param := e.Param(ParamKeyRoomID)
	roomID, err := strconv.ParseUint(param, 10, 64)
	if err != nil {
		return 0, NewHTTPError(http.StatusBadRequest, fmt.Errorf("requested room id(%v) is not allowed", param))
	}

	return roomID, nil
}

func (rest *RESTHandler) CreateRoom(e echo.Context) error {
	userID, ok := LoggedInUserID(e)
	if !ok {
		return ErrAPIRequireLoginFirst
	}

	createRoom := action.CreateRoom{}
	if err := e.Bind(&createRoom); err != nil {
		return err // default Bind returns *echo.NewHTTPError
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
	userID, ok := LoggedInUserID(e)
	if !ok {
		return ErrAPIRequireLoginFirst
	}
	roomID, err := validateParamRoomID(e)
	if err != nil {
		return err
	}

	deleteRoom := action.DeleteRoom{}
	deleteRoom.SenderID = userID
	deleteRoom.RoomID = roomID

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
	return e.JSON(http.StatusOK, response)
}

func (rest *RESTHandler) AddRoomMember(e echo.Context) error {
	userID, ok := LoggedInUserID(e)
	if !ok {
		return ErrAPIRequireLoginFirst
	}
	roomID, err := validateParamRoomID(e)
	if err != nil {
		return err
	}

	addRoomMember := action.AddRoomMember{}
	if err := e.Bind(&addRoomMember); err != nil {
		return err // default Bind returns *echo.NewHTTPError
	}
	addRoomMember.SenderID = userID
	addRoomMember.RoomID = roomID

	res, err := rest.chatCmd.AddRoomMember(e.Request().Context(), addRoomMember)
	if err != nil {
		if chat.IsNotFoundError(err) {
			return NewHTTPError(http.StatusNotFound, err)
		}
		return NewHTTPError(http.StatusInternalServerError, err)
	}

	response := struct {
		RoomID      uint64 `json:"added_room_id"`
		AddedUserID uint64 `json:"added_user_id"`
		OK          bool   `json:"ok"`
	}{
		RoomID:      res.RoomID,
		AddedUserID: res.UserID,
		OK:          true,
	}
	return e.JSON(http.StatusOK, response)
}

func (rest *RESTHandler) RemoveRoomMember(e echo.Context) error {
	userID, ok := LoggedInUserID(e)
	if !ok {
		return ErrAPIRequireLoginFirst
	}
	roomID, err := validateParamRoomID(e)
	if err != nil {
		return err
	}

	removeRoomMember := action.RemoveRoomMember{}
	if err := e.Bind(&removeRoomMember); err != nil {
		return err // default Bind returns *echo.NewHTTPError
	}
	removeRoomMember.SenderID = userID
	removeRoomMember.RoomID = roomID

	res, err := rest.chatCmd.RemoveRoomMember(e.Request().Context(), removeRoomMember)
	if err != nil {
		if chat.IsNotFoundError(err) {
			return NewHTTPError(http.StatusNotFound, err)
		}
		return NewHTTPError(http.StatusInternalServerError, err)
	}

	response := struct {
		RoomID        uint64 `json:"removed_room_id"`
		RemovedUserID uint64 `json:"removed_user_id"`
		OK            bool   `json:"ok"`
	}{
		RoomID:        res.RoomID,
		RemovedUserID: res.UserID,
		OK:            true,
	}
	return e.JSON(http.StatusOK, response)
}

func (rest *RESTHandler) GetRoomInfo(e echo.Context) error {
	userID, ok := LoggedInUserID(e)
	if !ok {
		return ErrAPIRequireLoginFirst
	}
	roomID, err := validateParamRoomID(e)
	if err != nil {
		return err
	}

	info, err := rest.chatQuery.FindRoomInfo(e.Request().Context(), userID, roomID)
	if err != nil {
		// TODO distinguish logic error or infra error
		// if _, ok := err.(*chat.InfraError); ok {
		return NewHTTPError(http.StatusInternalServerError, err)
	}

	return e.JSON(http.StatusOK, info)
}

func (rest *RESTHandler) GetUserInfo(e echo.Context) error {
	requestUserID, ok := LoggedInUserID(e)
	if !ok {
		return ErrAPIRequireLoginFirst
	}
	// TODO: use requestUserID to validate permittion for accessing other user info.
	_ = requestUserID

	queryUserID, err := validateParamUserID(e)
	if err != nil {
		return err
	}

	relation, err := rest.chatQuery.FindUserRelation(e.Request().Context(), queryUserID)
	if err != nil {
		return NewHTTPError(http.StatusInternalServerError, err)
	}

	return e.JSON(http.StatusOK, relation)
}

func (rest *RESTHandler) PostRoomMessage(e echo.Context) error {
	userID, ok := LoggedInUserID(e)
	if !ok {
		return ErrAPIRequireLoginFirst
	}
	roomID, err := validateParamRoomID(e)
	if err != nil {
		return err
	}

	postMsg := action.ChatMessage{}
	if err := e.Bind(&postMsg); err != nil {
		return err
	}
	postMsg.SenderID = userID
	postMsg.RoomID = roomID
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
	roomID, err := validateParamRoomID(e)
	if err != nil {
		return err
	}

	qRoomMsg := action.QueryRoomMessages{}
	if err := e.Bind(&qRoomMsg); err != nil {
		return err
	}
	qRoomMsg.RoomID = roomID

	roomMsg, err := rest.chatQuery.FindRoomMessages(e.Request().Context(), userID, qRoomMsg)
	if err != nil {
		return NewHTTPError(http.StatusInternalServerError, err)
	}

	return e.JSON(http.StatusOK, roomMsg)
}

func (rest *RESTHandler) GetUnreadRoomMessages(e echo.Context) error {
	userID, ok := LoggedInUserID(e)
	if !ok {
		return ErrAPIRequireLoginFirst
	}
	roomID, err := validateParamRoomID(e)
	if err != nil {
		return err
	}

	q := action.QueryUnreadRoomMessages{}
	if err := e.Bind(&q); err != nil {
		return err
	}
	q.RoomID = roomID

	unreads, err := rest.chatQuery.FindUnreadRoomMessages(e.Request().Context(), userID, q)
	if err != nil {
		return NewHTTPError(http.StatusInternalServerError, err)
	}

	return e.JSON(http.StatusOK, unreads)
}

func (rest *RESTHandler) ReadRoomMessages(e echo.Context) error {
	userID, ok := LoggedInUserID(e)
	if !ok {
		return ErrAPIRequireLoginFirst
	}
	roomID, err := validateParamRoomID(e)
	if err != nil {
		return err
	}

	readMessages := action.ReadMessages{}
	if err := e.Bind(&readMessages); err != nil {
		return err
	}
	readMessages.SenderID = userID
	readMessages.RoomID = roomID

	updatedRoomID, err := rest.chatCmd.ReadRoomMessages(e.Request().Context(), readMessages)
	if err != nil {
		return NewHTTPError(http.StatusInternalServerError, err)
	}

	response := struct {
		RoomID uint64 `json:"updated_room_id"`
		UserID uint64 `json:"read_user_id"`
		OK     bool   `json:"ok"`
	}{
		RoomID: updatedRoomID,
		UserID: readMessages.SenderID,
		OK:     true,
	}
	return e.JSON(http.StatusOK, response)
}
