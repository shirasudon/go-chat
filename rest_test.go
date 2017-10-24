package chat

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo"

	"github.com/shirasudon/go-chat/chat"
	"github.com/shirasudon/go-chat/chat/action"
	"github.com/shirasudon/go-chat/infra/pubsub"
)

func createRESTHandler() (rest *RESTHandler, doneFunc func()) {
	ps := pubsub.New(10)
	doneFunc = func() {
		ps.Shutdown()
	}
	return NewRESTHandler(
		NewLoginHandler(repository.Users()),
		chat.NewCommandService(repository, ps),
		chat.NewQueryService(repository),
	), doneFunc
}

const (
	createOrDeleteRoomID   = uint64(4)
	createOrDeleteByUserID = uint64(2)
)

func TestRESTCreateRoom(t *testing.T) {
	RESTHandler, done := createRESTHandler()
	defer done()

	createRoom := action.CreateRoom{}
	createRoom.ActionName = action.ActionCreateRoom
	createRoom.RoomName = "room1"
	createRoom.RoomMemberIDs = []uint64{2, 3}

	body, err := json.Marshal(createRoom)
	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(echo.POST, "/users/:id/rooms/new", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()

	c := theEcho.NewContext(req, rec)
	c.Set(KeyLoggedInUserID, createOrDeleteByUserID)
	c.SetParamNames("id")
	c.SetParamValues(fmt.Sprint(createOrDeleteByUserID))

	err = RESTHandler.CreateRoom(c)
	if err != nil {
		t.Fatal(err)
	}
	if expect, got := http.StatusCreated, rec.Code; expect != got {
		t.Errorf("different http status code, expect: %v, got: %v", expect, got)
	}

	response := make(map[string]interface{})
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	if err != nil {
		t.Fatal(err)
	}

	if roomID := uint64(response["room_id"].(float64)); roomID == 0 {
		t.Errorf("room created but created id is invalid")
	}
	if ok, assertionOK := response["ok"].(bool); !assertionOK || !ok {
		t.Errorf("room created but not ok status")
	}
	t.Logf("%#v", response)
}

func TestRESTDeleteRoom(t *testing.T) {
	RESTHandler, done := createRESTHandler()
	defer done()

	deleteRoom := action.DeleteRoom{}
	deleteRoom.ActionName = action.ActionDeleteRoom
	deleteRoom.RoomID = createOrDeleteRoomID

	body, err := json.Marshal(deleteRoom)
	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(echo.POST, "/rooms/delete", bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()

	c := theEcho.NewContext(req, rec)
	c.Set(KeyLoggedInUserID, createOrDeleteByUserID)
	c.SetParamNames("id")
	c.SetParamValues(fmt.Sprint(createOrDeleteByUserID))

	err = RESTHandler.DeleteRoom(c)
	if err != nil {
		t.Fatal(err)
	}
	if expect, got := http.StatusNoContent, rec.Code; expect != got {
		t.Errorf("different http status code, expect: %v, got: %v", expect, got)
	}

	response := make(map[string]interface{})
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	if err != nil {
		t.Fatal(err)
	}

	if roomID := uint64(response["room_id"].(float64)); roomID == 0 {
		t.Errorf("room deleted but deleted id is invalid")
	}
	if ok, assertionOK := response["ok"].(bool); !assertionOK || !ok {
		t.Errorf("room deleted but not ok status")
	}
	t.Logf("%#v", response)
}

func TestRESTGetUserRoom(t *testing.T) {
	req := httptest.NewRequest(echo.GET, "/users/1/rooms", nil)
	rec := httptest.NewRecorder()
	_ = theEcho.NewContext(req, rec)
}
