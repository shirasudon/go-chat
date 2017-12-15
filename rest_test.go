package chat

import (
	"bytes"
	"context"
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
	createRoom.RoomName = "room1"
	createRoom.RoomMemberIDs = []uint64{2, 3}

	req, err := newJSONRequest(echo.POST, "/rooms", createRoom)
	if err != nil {
		t.Fatal(err)
	}

	rec := httptest.NewRecorder()

	c := theEcho.NewContext(req, rec)
	c.Set(KeyLoggedInUserID, createOrDeleteByUserID)

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

// helper function for the echo.HTTPError.
func testAssertHTTPError(t *testing.T, target error, expectCode int, msgExists bool) {
	t.Helper()

	he, ok := target.(*echo.HTTPError)
	if !ok {
		t.Fatalf("invalid error type. expect: *echo.HTTPError, got: %#v", target)
	}
	if he.Code != expectCode {
		t.Errorf("different http status code for HTTPError, expect: %v, got: %v", expectCode, he.Code)
	}
	if msgExists && he.Message == nil {
		t.Errorf("error message is empty")
	}
}

func TestRESTCreateRoomFail(t *testing.T) {
	RESTHandler, done := createRESTHandler()
	defer done()

	// case 1: invalid request json
	{
		req := httptest.NewRequest(echo.POST, "/rooms", nil) // no json data and header
		rec := httptest.NewRecorder()

		c := theEcho.NewContext(req, rec)
		c.Set(KeyLoggedInUserID, createOrDeleteByUserID)

		err := RESTHandler.CreateRoom(c)
		if err == nil {
			t.Fatal("invalid json request is sent, but no error")
		}
		testAssertHTTPError(t, err, http.StatusBadRequest, true)
	}

	// case 2: referring not found user
	{
		const (
			NotFoundUserID = uint64(99)
		)
		createRoom := action.CreateRoom{}
		createRoom.RoomName = "room1"
		createRoom.RoomMemberIDs = []uint64{2, 3}

		req, err := newJSONRequest(echo.POST, "/rooms", createRoom)
		if err != nil {
			t.Fatal(err)
		}
		rec := httptest.NewRecorder()

		c := theEcho.NewContext(req, rec)
		c.Set(KeyLoggedInUserID, NotFoundUserID)

		err = RESTHandler.CreateRoom(c)
		if err == nil {
			t.Fatal("requesting not found user, but no error")
		}
		testAssertHTTPError(t, err, http.StatusInternalServerError, true)
	}
}

func TestRESTDeleteRoom(t *testing.T) {
	RESTHandler, done := createRESTHandler()
	defer done()

	req := httptest.NewRequest(echo.POST, "/rooms/:room_id", nil)
	rec := httptest.NewRecorder()

	c := theEcho.NewContext(req, rec)
	c.Set(KeyLoggedInUserID, createOrDeleteByUserID)
	c.SetParamNames("room_id")
	c.SetParamValues(fmt.Sprint(createOrDeleteRoomID))

	err := RESTHandler.DeleteRoom(c)
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

func TestRESTDeleteRoomFail(t *testing.T) {
	RESTHandler, done := createRESTHandler()
	defer done()

	// case 1: invalid request for resource parameter
	{
		req := httptest.NewRequest(echo.POST, "/rooms/:room_id", nil)
		rec := httptest.NewRecorder()

		c := theEcho.NewContext(req, rec)
		c.Set(KeyLoggedInUserID, createOrDeleteByUserID)
		c.SetParamNames("room_id")
		c.SetParamValues("room_id_string")

		err := RESTHandler.DeleteRoom(c)
		if err == nil {
			t.Fatal("invalid resource parameter is sent, but no error")
		}
		testAssertHTTPError(t, err, http.StatusBadRequest, true)
	}

	// case 2: referring not found user
	{
		const (
			NotFoundUserID = uint64(99)
		)
		req := httptest.NewRequest(echo.POST, "/rooms/:room_id", nil)
		rec := httptest.NewRecorder()

		c := theEcho.NewContext(req, rec)
		c.Set(KeyLoggedInUserID, NotFoundUserID)
		c.SetParamNames("room_id")
		c.SetParamValues(fmt.Sprint(createOrDeleteRoomID))

		err := RESTHandler.DeleteRoom(c)
		if err == nil {
			t.Fatal("requesting not found user, but no error")
		}
		testAssertHTTPError(t, err, http.StatusInternalServerError, true)
	}

	// case 3: referring not found room
	{
		const (
			NotFoundRoomID = uint64(99)
		)
		req := httptest.NewRequest(echo.POST, "/rooms/:room_id", nil)
		rec := httptest.NewRecorder()

		c := theEcho.NewContext(req, rec)
		c.Set(KeyLoggedInUserID, createOrDeleteRoomID)
		c.SetParamNames("room_id")
		c.SetParamValues(fmt.Sprint(NotFoundRoomID))

		err := RESTHandler.DeleteRoom(c)
		if err == nil {
			t.Fatal("requesting not found room, but no error")
		}
		testAssertHTTPError(t, err, http.StatusInternalServerError, true)
	}
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

	// just check for existance of json keys.
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
	RESTHandler, done := createRESTHandler()
	defer done()

	// case 1: invalid request for resource parameter
	{
		req := httptest.NewRequest(echo.GET, "/rooms/:room_id", nil)
		rec := httptest.NewRecorder()

		c := theEcho.NewContext(req, rec)
		c.Set(KeyLoggedInUserID, createOrDeleteByUserID)
		c.SetParamNames("room_id")
		c.SetParamValues("room_id_string")

		err := RESTHandler.GetRoomInfo(c)
		if err == nil {
			t.Fatal("invalid resource parameter is sent, but no error")
		}
		testAssertHTTPError(t, err, http.StatusBadRequest, true)
	}

	// case 2: referring not found room
	{
		const (
			NotFoundRoomID = uint64(99)
			loginUserID    = uint64(2)
		)

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
}

func TestRESTGetUserInfo(t *testing.T) {
	RESTHandler, done := createRESTHandler()
	defer done()

	const (
		TestUserID = uint64(2)
	)

	req := httptest.NewRequest(echo.GET, "/users/2", nil)
	rec := httptest.NewRecorder()

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

func TestRESTGetUserInfoFail(t *testing.T) {
	RESTHandler, done := createRESTHandler()
	defer done()

	// case 1: invalid request for resource parameter
	{
		req := httptest.NewRequest(echo.GET, "/users/:user_id", nil)
		rec := httptest.NewRecorder()

		c := theEcho.NewContext(req, rec)
		c.Set(KeyLoggedInUserID, createOrDeleteByUserID)
		c.SetParamNames("user_id")
		c.SetParamValues("user_id_string")

		err := RESTHandler.GetUserInfo(c)
		if err == nil {
			t.Fatal("invalid resource parameter is sent, but no error")
		}
		testAssertHTTPError(t, err, http.StatusBadRequest, true)
	}

	// case 2: referring not found user
	{
		const (
			NotFoundUserID = uint64(99)
			loginUserID    = uint64(2)
		)

		req := httptest.NewRequest(echo.GET, "/users/:user_id", nil)
		rec := httptest.NewRecorder()

		c := theEcho.NewContext(req, rec)
		c.Set(KeyLoggedInUserID, loginUserID)
		c.SetParamNames("user_id")
		c.SetParamValues(fmt.Sprint(NotFoundUserID))

		err := RESTHandler.GetUserInfo(c)
		if err == nil {
			t.Fatal("not found user is specified but no error")
		}
	}
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

	// case 1: found resource
	{
		query := action.QueryRoomMessages{
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
	}

	// case 2: not found resource
	{
		const (
			UserID         = uint64(2)
			NotFoundRoomID = uint64(999)
		)

		query := action.QueryRoomMessages{
			Before: time.Now(),
			Limit:  1,
		}

		req, err := newJSONRequest(echo.GET, "/rooms/:room_id/messages", query)
		if err != nil {
			t.Fatal(err)
		}

		rec := httptest.NewRecorder()

		c := theEcho.NewContext(req, rec)
		c.Set(KeyLoggedInUserID, UserID)
		c.SetParamNames("room_id")
		c.SetParamValues(fmt.Sprint(NotFoundRoomID))

		err = RESTHandler.GetRoomMessages(c)
		if err == nil {
			t.Error("given not found room id but no error")
		}
	}
}

func TestRESTGetUnreadRoomMessages(t *testing.T) {
	RESTHandler := NewRESTHandler(
		chat.NewCommandService(repository, globalPubsub),
		chat.NewQueryService(queryers),
	)

	{ // create room messages
		chatMsg := action.ChatMessage{}
		chatMsg.RoomID = createMsgRoomID
		chatMsg.SenderID = createOrDeleteByUserID
		chatMsg.Content = "hello"
		_, err := RESTHandler.chatCmd.PostRoomMessage(context.Background(), chatMsg)
		if err != nil {
			t.Fatal(err)
		}
		time.Sleep(10 * time.Millisecond)
	}

	// case1: found
	{
		query := action.QueryUnreadRoomMessages{
			Limit: 1,
		}

		req, err := newJSONRequest(echo.GET, "/rooms/:room_id/messages/unread", query)
		if err != nil {
			t.Fatal(err)
		}

		rec := httptest.NewRecorder()

		c := theEcho.NewContext(req, rec)
		c.Set(KeyLoggedInUserID, createOrDeleteByUserID)
		c.SetParamNames("room_id")
		c.SetParamValues(fmt.Sprint(createMsgRoomID))

		err = RESTHandler.GetUnreadRoomMessages(c)
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

		// just check for existance of json keys.
		for _, key := range []string{
			"room_id",
			"messages",
			"messages_size",
		} {
			if _, ok := response[key]; !ok {
				t.Errorf("missing field (%v) in json response", key)
			}
		}
	}

	// case2: room id is not found
	{
		const (
			NotFoundRoomID = uint64(999)
			UserID         = uint64(2)
		)

		query := action.QueryUnreadRoomMessages{
			Limit: 1,
		}

		req, err := newJSONRequest(echo.GET, "/rooms/:room_id/messages/unread", query)
		if err != nil {
			t.Fatal(err)
		}

		rec := httptest.NewRecorder()

		c := theEcho.NewContext(req, rec)
		c.Set(KeyLoggedInUserID, UserID)
		c.SetParamNames("room_id")
		c.SetParamValues(fmt.Sprint(NotFoundRoomID))

		err = RESTHandler.GetUnreadRoomMessages(c)
		if err == nil {
			t.Error("given not found room id but no error")
		}
	}
}
