package chat

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

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
		chat.NewCommandService(repository, ps),
		chat.NewQueryService(queryers),
	), doneFunc
}

func newJSONRequest(method, url string, data interface{}) (*http.Request, error) {
	body, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	req := httptest.NewRequest(method, url, bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	return req, nil
}

const (
	createOrDeleteRoomID   = uint64(4)
	createOrDeleteByUserID = uint64(2)
	createMsgRoomID        = uint64(3)
	createMsgContent       = "hello!"
)

func TestRESTCreateRoom(t *testing.T) {
	RESTHandler, done := createRESTHandler()
	defer done()

	createRoom := action.CreateRoom{}
	createRoom.ActionName = action.ActionCreateRoom
	createRoom.RoomName = "room1"
	createRoom.RoomMemberIDs = []uint64{2, 3}

	req, err := newJSONRequest(echo.POST, "/users/:user_id/rooms/new", createRoom)
	if err != nil {
		t.Fatal(err)
	}

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

	req, err := newJSONRequest(echo.POST, "/users/:user_id/rooms/:room_id/delete", deleteRoom)
	if err != nil {
		t.Fatal(err)
	}

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

func TestRESTGetRoomInfoSuccess(t *testing.T) {
	const (
		getRoomID   = uint64(3)
		loginUserID = uint64(2)
	)
	RESTHandler, done := createRESTHandler()
	defer done()

	req := httptest.NewRequest(echo.GET, "/rooms/:room_id", nil)
	rec := httptest.NewRecorder()

	c := theEcho.NewContext(req, rec)
	c.Set(KeyLoggedInUserID, loginUserID)
	c.SetParamNames("room_id")
	c.SetParamValues(fmt.Sprint(getRoomID))

	err := RESTHandler.GetRoomInfo(c)
	if err != nil {
		t.Fatal(err)
	}
	if expect, got := http.StatusFound, rec.Code; expect != got {
		t.Errorf("different http status code, expect: %v, got: %v", expect, got)
	}

	response := make(map[string]interface{})
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	if err != nil {
		t.Fatal(err)
	}

	for _, key := range []string{
		"room_name",
		"room_creator_id",
		"room_members",
		"room_members_size",
	} {
		if _, ok := response[key]; !ok {
			t.Errorf("missing field (%v) in json response", key)
		}
	}

	members, ok := response["room_members"].([]interface{})
	if !ok {
		t.Fatalf("The response has invalid field (%v); value: %#v", "room_members", response)
	}

	for _, member := range members {
		for _, key := range []string{
			"user_name",
			"user_id",
			"first_name",
			"last_name",
		} {
			if _, ok := member.(map[string]interface{})[key]; !ok {
				t.Errorf("missing field (%v) in room member", key)
			}
		}
	}
}

func TestRESTGetRoomInfoFail(t *testing.T) {
	const (
		NotFoundRoomID = uint64(99)
		loginUserID    = uint64(2)
	)
	RESTHandler, done := createRESTHandler()
	defer done()

	req := httptest.NewRequest(echo.GET, "/rooms/:room_id", nil)
	rec := httptest.NewRecorder()

	c := theEcho.NewContext(req, rec)
	c.Set(KeyLoggedInUserID, loginUserID)
	c.SetParamNames("room_id")
	c.SetParamValues(fmt.Sprint(NotFoundRoomID))

	err := RESTHandler.GetRoomInfo(c)
	if err == nil {
		t.Fatal("not found room is specified but no error")
	}
}

func TestRESTGetUserInfo(t *testing.T) {
	RESTHandler, done := createRESTHandler()
	defer done()

	req := httptest.NewRequest(echo.GET, "/users/2", nil)
	rec := httptest.NewRecorder()

	const TestUserID = uint64(2)
	c := theEcho.NewContext(req, rec)
	c.Set(KeyLoggedInUserID, TestUserID)
	c.SetParamNames("user_id")
	c.SetParamValues(fmt.Sprint(TestUserID))

	err := RESTHandler.GetUserInfo(c)
	if err != nil {
		t.Fatal(err)
	}
	if expect, got := http.StatusFound, rec.Code; expect != got {
		t.Errorf("different http status code, expect: %v, got: %v", expect, got)
	}

	response := make(map[string]interface{})
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	if err != nil {
		t.Fatal(err)
	}

	if userID := uint64(response["user_id"].(float64)); userID != TestUserID {
		t.Errorf("returning different user id, expect: %d, got: %d", TestUserID, userID)
	}
	t.Logf("Response: %#v", response)
}

func TestRESTPostRoomMessage(t *testing.T) {
	RESTHandler, done := createRESTHandler()
	defer done()

	chatMsg := action.ChatMessage{}
	chatMsg.Content = createMsgContent

	req, err := newJSONRequest(echo.POST, "/rooms/:room_id/messages", chatMsg)
	if err != nil {
		t.Fatal(err)
	}

	rec := httptest.NewRecorder()

	c := theEcho.NewContext(req, rec)
	c.Set(KeyLoggedInUserID, createOrDeleteByUserID)
	c.SetParamNames("room_id")
	c.SetParamValues(fmt.Sprint(createMsgRoomID))

	err = RESTHandler.PostRoomMessage(c)
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

	if msgID := uint64(response["message_id"].(float64)); msgID == 0 {
		t.Errorf("message created but created message id is invalid")
	}
	if roomID := uint64(response["room_id"].(float64)); roomID == 0 {
		t.Errorf("message created but target room id is invalid")
	}
	if ok, assertionOK := response["ok"].(bool); !assertionOK || !ok {
		t.Errorf("message created but not ok status")
	}
	t.Logf("%#v", response)
}

func TestRESTGetRoomMessages(t *testing.T) {
	RESTHandler, done := createRESTHandler()
	defer done()

	query := action.QueryRoomMessages{
		RoomID: createMsgRoomID,
		Before: time.Now(),
		Limit:  1,
	}

	req, err := newJSONRequest(echo.GET, "/rooms/:room_id/messages", query)
	if err != nil {
		t.Fatal(err)
	}

	rec := httptest.NewRecorder()

	c := theEcho.NewContext(req, rec)
	c.Set(KeyLoggedInUserID, createOrDeleteByUserID)
	c.SetParamNames("room_id")
	c.SetParamValues(fmt.Sprint(createMsgRoomID))

	err = RESTHandler.GetRoomMessages(c)
	if err != nil {
		t.Fatal(err)
	}
	if expect, got := http.StatusFound, rec.Code; expect != got {
		t.Errorf("different http status code, expect: %v, got: %v", expect, got)
	}

	response := make(map[string]interface{})
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	if err != nil {
		t.Fatal(err)
	}

	msgs, ok := response["messages"].([]interface{})
	if !ok {
		t.Fatalf("response has invalid structure, %#v", response)
	}
	if len(msgs) != 1 {
		t.Fatalf("different messages size, expect: %v, got: %v", 1, len(msgs))
	}
	msg, ok := msgs[0].(map[string]interface{})
	if !ok {
		t.Fatalf("response.messages has invalid structure, %#v", msgs)
	}
	if expect, got := createMsgContent, msg["content"]; expect != got {
		t.Errorf("different message content, expect: %v, got: %v", expect, got)
	}

	if roomID := uint64(response["room_id"].(float64)); roomID == 0 {
		t.Errorf("message created but target room id is invalid")
	}

	t.Logf("%#v", response)
}
