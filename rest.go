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
	loginHandler *LoginHandler
	cmdService   *chat.CommandService
	queryService *chat.QueryService
}

func NewRESTHandler(loginHandler *LoginHandler, cmdService *chat.CommandService, queryService *chat.QueryService) *RESTHandler {
	return &RESTHandler{
		loginHandler: loginHandler,
		cmdService:   cmdService,
		queryService: queryService,
	}
}

func (rest *RESTHandler) RegisterPath(g *echo.Group) {
	g.POST("/users/:id/rooms", rest.CreateRoom)
	g.DELETE("/users/:id/rooms", rest.DeleteRoom)
}

func (rest *RESTHandler) validateUserID(e echo.Context) (uint64, error) {
	userID, ok := rest.loginHandler.LoggedInUserID(e)
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

	createdID, err := rest.cmdService.CreateRoom(e.Request().Context(), createRoom)
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

	deletedID, err := rest.cmdService.DeleteRoom(e.Request().Context(), deleteRoom)
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
	userID, ok := rest.loginHandler.LoggedInUserID(e)
	if !ok {
		return ErrRequireLoginFirst
	}

	rooms, err := rest.queryService.FindUserRooms(e.Request().Context(), userID)
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
