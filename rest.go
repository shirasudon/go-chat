package chat

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/labstack/echo"

	"github.com/shirasudon/go-chat/chat"
	"github.com/shirasudon/go-chat/chat/action"
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

func (rest *RESTHandler) validateUserID(e echo.Context) (uint64, error) {
	userID, ok := LoggedInUserID(e)
	if !ok {
		return 0, echo.NewHTTPError(http.StatusUnauthorized, ErrRequireLoginFirst)
	}

	userPathID, err := strconv.ParseUint(e.Param("id"), 10, 64)
	if err != nil || userPathID != userID {
		return 0, echo.NewHTTPError(http.StatusBadRequest, fmt.Errorf("requested user id(%v) is not allowed", userPathID))
	}

	return userID, nil
}

func (rest *RESTHandler) CreateRoom(e echo.Context) error {
	userID, err := rest.validateUserID(e)
	if err != nil {
		return err
	}

	createRoom := action.CreateRoom{}
	if err = e.Bind(&createRoom); err != nil {
		// TODO return error as JSON format
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}
	createRoom.SenderID = userID

	createdID, err := rest.chatCmd.CreateRoom(e.Request().Context(), createRoom)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
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
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}
	deleteRoom.SenderID = userID

	deletedID, err := rest.chatCmd.DeleteRoom(e.Request().Context(), deleteRoom)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
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

func (rest *RESTHandler) GetUserRooms(e echo.Context) error {
	userID, ok := LoggedInUserID(e)
	if !ok {
		return ErrRequireLoginFirst
	}

	rooms, err := rest.chatQuery.FindUserRooms(e.Request().Context(), userID)
	if err != nil {
		return err
	}

	// TODO move to query service
	type userRoom struct {
		RoomName string `json:"room_name"`
		OwnerID  uint64 `json:"owner_id"`
		Members  []struct {
			UserID   uint64 `json:"user_id"`
			UserName string `json:"user_name"`
		} `json:"members"`
	}
	userRooms := make([]userRoom, 0, len(rooms))
	for _, r := range rooms {
		ur := userRoom{
			RoomName: r.Name,
			OwnerID:  userID,
		}
		// TODO get user information
		// ur.Members = ...
		userRooms = append(userRooms, ur)
	}

	return e.JSON(http.StatusOK, userRooms)
}
