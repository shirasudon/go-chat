package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/labstack/echo"

	"github.com/golang/mock/gomock"

	"github.com/shirasudon/go-chat/chat"
	"github.com/shirasudon/go-chat/chat/action"
	"github.com/shirasudon/go-chat/chat/queried"
	"github.com/shirasudon/go-chat/chat/result"
	"github.com/shirasudon/go-chat/infra/pubsub"
	"github.com/shirasudon/go-chat/internal/mocks"
)

func createRESTHandler() (rest *RESTHandler, doneFunc func()) {
	ps := pubsub.New(10)
	doneFunc = func() {
		ps.Shutdown()
	}
	return NewRESTHandler(
		chat.NewCommandServiceImpl(repository, ps),
		chat.NewQueryServiceImpl(queryers),
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

func TestRESTRequireLoggeinUserID(t *testing.T) {
	RESTHandler, done := createRESTHandler()
	defer done()

	handlers := []struct {
		Name    string
		Handler echo.HandlerFunc
	}{
		{"CreateRoom", RESTHandler.CreateRoom},
		{"DeleteRoom", RESTHandler.DeleteRoom},
		{"AddRoomMember", RESTHandler.AddRoomMember},
		{"RemoveRoomMember", RESTHandler.RemoveRoomMember},
		{"GetRoomInfo", RESTHandler.GetRoomInfo},
		{"GetUserInfo", RESTHandler.GetUserInfo},
		{"PostRoomMessage", RESTHandler.PostRoomMessage},
		{"GetRoomMessages", RESTHandler.GetRoomMessages},
		{"GetUnreadRoomMessages", RESTHandler.GetUnreadRoomMessages},
		{"ReadRoomMessages", RESTHandler.ReadRoomMessages},
	}

	// expect to return LoginError.
	for _, h := range handlers {
		req := httptest.NewRequest(echo.GET, "/", nil)
		rec := httptest.NewRecorder()
		c := theEcho.NewContext(req, rec)

		if err := h.Handler(c); err != ErrAPIRequireLoginFirst {
			t.Errorf("%v: requesting with not loggedin state but login error is not returned: %v", h.Name, err)
		}
	}

	// expect to return not LoginError.
	for _, h := range handlers {
		req := httptest.NewRequest(echo.GET, "/", nil)
		rec := httptest.NewRecorder()
		c := theEcho.NewContext(req, rec)
		c.Set(KeyLoggedInUserID, uint64(1))

		if err := h.Handler(c); err == ErrAPIRequireLoginFirst {
			t.Errorf("%v: requesting with loggedin state but login error is returned: %v", h.Name, err)
		}
	}
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
	if expect, got := http.StatusOK, rec.Code; expect != got {
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

func TestRESTAddRoomMember(t *testing.T) {
	t.Parallel()

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	// case 1: success
	{
		const (
			UserID    = uint64(1)
			RoomID    = uint64(2)
			AddUserID = uint64(3)
		)
		AddRoomMember := action.AddRoomMember{
			SenderID:  UserID,
			RoomID:    RoomID,
			AddUserID: AddUserID,
		}

		cmdService := mocks.NewMockCommandService(mockCtrl)
		cmdService.EXPECT().AddRoomMember(gomock.Any(), AddRoomMember).
			Return(&result.AddRoomMember{
				RoomID: RoomID,
				UserID: UserID,
			}, nil).Times(1)
		RESTHandler := &RESTHandler{chatCmd: cmdService}

		req, err := newJSONRequest(echo.POST, "/rooms/:room_id/members", AddRoomMember)
		if err != nil {
			t.Fatal(err)
		}
		rec := httptest.NewRecorder()

		c := theEcho.NewContext(req, rec)
		c.Set(KeyLoggedInUserID, UserID)
		c.SetParamNames("room_id")
		c.SetParamValues(fmt.Sprint(RoomID))

		err = RESTHandler.AddRoomMember(c)
		if err != nil {
			t.Fatalf("AddRoomMember returns error: %v", err)
		}

		response := make(map[string]interface{})
		err = json.Unmarshal(rec.Body.Bytes(), &response)
		if err != nil {
			t.Fatal(err)
		}

		for _, expect := range []struct {
			Key   string
			Value interface{}
			Equal func(x, y interface{}) bool
		}{
			{"added_room_id", RoomID, func(x, y interface{}) bool { return x.(uint64) == uint64(y.(float64)) }},
			{"added_user_id", UserID, func(x, y interface{}) bool { return x.(uint64) == uint64(y.(float64)) }},
			{"ok", true, func(x, y interface{}) bool { return x.(bool) == y.(bool) }},
		} {
			if !expect.Equal(expect.Value, response[expect.Key]) {
				t.Errorf("differenct value for key(%v), expect: %v, got: %v", expect.Key, expect.Value, response[expect.Key])
			}
		}
	}

	// case 2: referring not found resource
	{
		const (
			NotFoundUserID = uint64(1)
			NotFoundRoomID = uint64(2)
		)
		AddRoomMember := action.AddRoomMember{
			SenderID:  NotFoundUserID,
			RoomID:    NotFoundRoomID,
			AddUserID: NotFoundUserID,
		}

		cmdService := mocks.NewMockCommandService(mockCtrl)
		cmdService.EXPECT().AddRoomMember(gomock.Any(), AddRoomMember).
			Return(nil, chat.NewNotFoundError("not found")).Times(1)
		RESTHandler := &RESTHandler{chatCmd: cmdService}

		req, err := newJSONRequest(echo.POST, "/rooms/:room_id/members", AddRoomMember)
		if err != nil {
			t.Fatal(err)
		}
		rec := httptest.NewRecorder()

		c := theEcho.NewContext(req, rec)
		c.Set(KeyLoggedInUserID, NotFoundUserID)
		c.SetParamNames("room_id")
		c.SetParamValues(fmt.Sprint(NotFoundRoomID))

		err = RESTHandler.AddRoomMember(c)
		if err == nil {
			t.Fatal("requesting not found user and room, but no error")
		}
		testAssertHTTPError(t, err, http.StatusNotFound, true)
	}

	// case 3: bad request
	{
		const (
			UserID = uint64(1)
		)
		AddRoomMember := action.AddRoomMember{}

		cmdService := mocks.NewMockCommandService(mockCtrl)
		RESTHandler := &RESTHandler{chatCmd: cmdService}

		req, err := newJSONRequest(echo.POST, "/rooms/:room_id/members", AddRoomMember)
		if err != nil {
			t.Fatal(err)
		}
		rec := httptest.NewRecorder()

		c := theEcho.NewContext(req, rec)
		c.Set(KeyLoggedInUserID, UserID)
		c.SetParamNames("room_id")
		c.SetParamValues("invalid_room_id")

		err = RESTHandler.AddRoomMember(c)
		if err == nil {
			t.Fatal("requesting bad arguments, but no error")
		}
		testAssertHTTPError(t, err, http.StatusBadRequest, true)
	}
}

func TestRESTRemoveRoomMember(t *testing.T) {
	t.Parallel()

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	// case 1: success
	{
		const (
			UserID       = uint64(1)
			RoomID       = uint64(2)
			RemoveUserID = uint64(3)
		)
		RemoveRoomMember := action.RemoveRoomMember{
			SenderID:     UserID,
			RoomID:       RoomID,
			RemoveUserID: RemoveUserID,
		}

		cmdService := mocks.NewMockCommandService(mockCtrl)
		cmdService.EXPECT().RemoveRoomMember(gomock.Any(), RemoveRoomMember).
			Return(&result.RemoveRoomMember{
				RoomID: RoomID,
				UserID: UserID,
			}, nil).Times(1)
		RESTHandler := &RESTHandler{chatCmd: cmdService}

		req, err := newJSONRequest(echo.DELETE, "/rooms/:room_id/members", RemoveRoomMember)
		if err != nil {
			t.Fatal(err)
		}
		rec := httptest.NewRecorder()

		c := theEcho.NewContext(req, rec)
		c.Set(KeyLoggedInUserID, UserID)
		c.SetParamNames("room_id")
		c.SetParamValues(fmt.Sprint(RoomID))

		err = RESTHandler.RemoveRoomMember(c)
		if err != nil {
			t.Fatalf("RemoveRoomMember returns error: %v", err)
		}

		response := make(map[string]interface{})
		err = json.Unmarshal(rec.Body.Bytes(), &response)
		if err != nil {
			t.Fatal(err)
		}

		for _, expect := range []struct {
			Key   string
			Value interface{}
			Equal func(x, y interface{}) bool
		}{
			{"removed_room_id", RoomID, func(x, y interface{}) bool { return x.(uint64) == uint64(y.(float64)) }},
			{"removed_user_id", UserID, func(x, y interface{}) bool { return x.(uint64) == uint64(y.(float64)) }},
			{"ok", true, func(x, y interface{}) bool { return x.(bool) == y.(bool) }},
		} {
			if !expect.Equal(expect.Value, response[expect.Key]) {
				t.Errorf("differenct value for key(%v), expect: %v, got: %v", expect.Key, expect.Value, response[expect.Key])
			}
		}
	}

	// case 2: referring not found resource
	{
		const (
			NotFoundUserID = uint64(1)
			NotFoundRoomID = uint64(2)
		)
		RemoveRoomMember := action.RemoveRoomMember{
			SenderID:     NotFoundUserID,
			RoomID:       NotFoundRoomID,
			RemoveUserID: NotFoundUserID,
		}

		cmdService := mocks.NewMockCommandService(mockCtrl)
		cmdService.EXPECT().RemoveRoomMember(gomock.Any(), RemoveRoomMember).
			Return(nil, chat.NewNotFoundError("not found")).Times(1)
		RESTHandler := &RESTHandler{chatCmd: cmdService}

		req, err := newJSONRequest(echo.DELETE, "/rooms/:room_id/members", RemoveRoomMember)
		if err != nil {
			t.Fatal(err)
		}
		rec := httptest.NewRecorder()

		c := theEcho.NewContext(req, rec)
		c.Set(KeyLoggedInUserID, NotFoundUserID)
		c.SetParamNames("room_id")
		c.SetParamValues(fmt.Sprint(NotFoundRoomID))

		err = RESTHandler.RemoveRoomMember(c)
		if err == nil {
			t.Fatal("requesting not found user and room, but no error")
		}
		testAssertHTTPError(t, err, http.StatusNotFound, true)
	}

	// case 3: bad request
	{
		const (
			UserID = uint64(1)
		)
		RemoveRoomMember := action.RemoveRoomMember{}

		cmdService := mocks.NewMockCommandService(mockCtrl)
		RESTHandler := &RESTHandler{chatCmd: cmdService}

		req, err := newJSONRequest(echo.DELETE, "/rooms/:room_id/members", RemoveRoomMember)
		if err != nil {
			t.Fatal(err)
		}
		rec := httptest.NewRecorder()

		c := theEcho.NewContext(req, rec)
		c.Set(KeyLoggedInUserID, UserID)
		c.SetParamNames("room_id")
		c.SetParamValues("invalid_room_id")

		err = RESTHandler.RemoveRoomMember(c)
		if err == nil {
			t.Fatal("requesting bad arguments, but no error")
		}
		testAssertHTTPError(t, err, http.StatusBadRequest, true)
	}
}

// TODO rewrite test function with gomock, which other tests depends on command service

func TestRESTReadRoomMessages(t *testing.T) {
	t.Parallel()

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	{
		// case 1: success
		var (
			ReadMessages = action.ReadMessages{
				ReadAt: time.Now(),
			}
		)

		cmdService := mocks.NewMockCommandService(mockCtrl)
		cmdService.EXPECT().
			ReadRoomMessages(gomock.Any(), gomock.Not(action.ReadMessages{ReadAt: time.Time{}})).
			Return(ReadMessages.RoomID, nil).
			Times(1)

		req, err := newJSONRequest(echo.POST, "/rooms/:room_id/messages/read", ReadMessages)
		if err != nil {
			t.Fatal(err)
		}
		rec := httptest.NewRecorder()

		c := theEcho.NewContext(req, rec)
		c.Set(KeyLoggedInUserID, ReadMessages.SenderID)
		c.SetParamNames("room_id")
		c.SetParamValues(fmt.Sprint(ReadMessages.RoomID))

		RESTHandler := &RESTHandler{chatCmd: cmdService}
		err = RESTHandler.ReadRoomMessages(c)
		if err != nil {
			t.Fatal(err)
		}
		if expect, got := http.StatusOK, rec.Code; expect != got {
			t.Errorf("different http status code, expect: %v, got: %v", expect, got)
		}

		response := make(map[string]interface{})
		err = json.Unmarshal(rec.Body.Bytes(), &response)
		if err != nil {
			t.Fatal(err)
		}

		if roomID := uint64(response["updated_room_id"].(float64)); roomID != ReadMessages.RoomID {
			t.Errorf("returned room ID is invalid, expect: %v, got: %v", ReadMessages.RoomID, roomID)
		}
		if userID := uint64(response["read_user_id"].(float64)); userID != ReadMessages.SenderID {
			t.Errorf("returned user ID is invalid, expect: %v, got: %v", ReadMessages.SenderID, userID)
		}
		if ok, assertionOK := response["ok"].(bool); !assertionOK || !ok {
			t.Errorf("read room messages succeeded but not ok status")
		}
	}
}

func TestRESTGetRoomInfoSuccess(t *testing.T) {
	t.Parallel()

	const (
		getRoomID   = uint64(3)
		loginUserID = uint64(2)
	)
	RESTHandler, done := createRESTHandler()
	defer done()

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	var emptyMsgs = queried.EmptyRoomInfo
	emptyMsgs.RoomID = getRoomID

	qs := mocks.NewMockQueryService(mockCtrl)
	qs.EXPECT().
		FindRoomInfo(gomock.Any(), loginUserID, getRoomID).
		Return(&emptyMsgs, nil).
		Times(1)

	RESTHandler.chatQuery = qs

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
	if expect, got := http.StatusOK, rec.Code; expect != got {
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
			"message_read_at",
		} {
			if _, ok := member.(map[string]interface{})[key]; !ok {
				t.Errorf("missing field (%v) in room member", key)
			}
		}
	}
}

func TestRESTGetRoomInfoFail(t *testing.T) {
	t.Parallel()

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	// case 1: invalid request for resource parameter
	{
		const (
			loginUserID = uint64(2)
		)

		qs := mocks.NewMockQueryService(mockCtrl)
		qs.EXPECT().
			FindRoomInfo(gomock.Any(), gomock.Any(), gomock.Any()).
			Times(0)

		RESTHandler := &RESTHandler{chatQuery: qs}

		req := httptest.NewRequest(echo.GET, "/rooms/:room_id", nil)
		rec := httptest.NewRecorder()

		c := theEcho.NewContext(req, rec)
		c.Set(KeyLoggedInUserID, loginUserID)
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

		qs := mocks.NewMockQueryService(mockCtrl)
		qs.EXPECT().
			FindRoomInfo(gomock.Any(), loginUserID, NotFoundRoomID).
			Return(nil, chat.NewNotFoundError("not found")).
			Times(1)

		RESTHandler := &RESTHandler{chatQuery: qs}

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
	t.Parallel()

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	const (
		TestUserID = uint64(2)
	)

	ur := queried.EmptyUserRelation
	ur.UserID = TestUserID

	qs := mocks.NewMockQueryService(mockCtrl)
	qs.EXPECT().
		FindUserRelation(gomock.Any(), TestUserID).
		Return(&ur, nil).
		Times(1)

	RESTHandler := &RESTHandler{chatQuery: qs}

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
	if expect, got := http.StatusOK, rec.Code; expect != got {
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
	t.Parallel()

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	// case 1: invalid request for resource parameter
	{
		const (
			TestUserID = uint64(2)
		)
		qs := mocks.NewMockQueryService(mockCtrl)
		qs.EXPECT().FindUserRelation(gomock.Any(), gomock.Any()).Times(0)

		RESTHandler := &RESTHandler{chatQuery: qs}

		req := httptest.NewRequest(echo.GET, "/users/:user_id", nil)
		rec := httptest.NewRecorder()

		c := theEcho.NewContext(req, rec)
		c.Set(KeyLoggedInUserID, TestUserID)
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

		qs := mocks.NewMockQueryService(mockCtrl)
		qs.EXPECT().
			FindUserRelation(gomock.Any(), NotFoundUserID).
			Return(nil, chat.NewNotFoundError("")).
			Times(1)

		RESTHandler := &RESTHandler{chatQuery: qs}

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
	const URL = "/rooms/:room_id/messages"

	t.Parallel()

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	// case: found resource with json
	{
		const (
			LoginUserID = uint64(1)
			RoomID      = uint64(2)
		)

		res := queried.EmptyRoomMessages
		res.RoomID = RoomID
		res.Msgs = []queried.Message{queried.Message{}}

		query := action.QueryRoomMessages{
			RoomID: RoomID,
			Before: NormTimestampNow(),
			Limit:  1,
		}

		qs := mocks.NewMockQueryService(mockCtrl)
		qs.EXPECT().
			FindRoomMessages(gomock.Any(), LoginUserID, query).
			Return(&res, nil).
			Times(1)

		RESTHandler := &RESTHandler{chatQuery: qs}

		query.RoomID = 0 // remove RoomID from query JSON.
		req, err := newJSONRequest(echo.GET, URL, query)
		if err != nil {
			t.Fatal(err)
		}
		rec := httptest.NewRecorder()

		c := theEcho.NewContext(req, rec)
		c.Set(KeyLoggedInUserID, LoginUserID)
		c.SetParamNames("room_id")
		c.SetParamValues(fmt.Sprint(RoomID))

		err = RESTHandler.GetRoomMessages(c)
		if err != nil {
			t.Fatal(err)
		}
		if expect, got := http.StatusOK, rec.Code; expect != got {
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
		if _, ok := msgs[0].(map[string]interface{}); !ok {
			t.Fatalf("response.messages has invalid structure, %#v", msgs)
		}

		if roomID := uint64(response["room_id"].(float64)); roomID != RoomID {
			t.Errorf("message created but target room id is invalid")
		}
	}

	// case: found resource with query parameter
	{
		const (
			LoginUserID = uint64(1)
			RoomID      = uint64(2)
		)

		res := queried.EmptyRoomMessages
		res.RoomID = RoomID
		res.Msgs = []queried.Message{queried.Message{}}

		actionQuery := action.QueryRoomMessages{
			RoomID: RoomID,
			Before: NormTimestampNow(),
			Limit:  1,
		}

		qs := mocks.NewMockQueryService(mockCtrl)
		qs.EXPECT().
			FindRoomMessages(gomock.Any(), LoginUserID, actionQuery).
			Return(&res, nil).
			Times(1)

		RESTHandler := &RESTHandler{chatQuery: qs}

		byteTime := MustMarshal(actionQuery.Before.MarshalText())
		query := make(url.Values)
		query.Set("before", string(byteTime))
		query.Set("limit", fmt.Sprint(actionQuery.Limit))
		req := httptest.NewRequest(echo.GET, URL+"/?"+query.Encode(), nil)
		rec := httptest.NewRecorder()

		c := theEcho.NewContext(req, rec)
		c.Set(KeyLoggedInUserID, LoginUserID)
		c.SetParamNames("room_id")
		c.SetParamValues(fmt.Sprint(RoomID))

		t.Log(c.QueryParam("before"), string(byteTime))

		err := RESTHandler.GetRoomMessages(c)
		if err != nil {
			t.Fatal(err)
		}
		if expect, got := http.StatusOK, rec.Code; expect != got {
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
		if _, ok := msgs[0].(map[string]interface{}); !ok {
			t.Fatalf("response.messages has invalid structure, %#v", msgs)
		}

		if roomID := uint64(response["room_id"].(float64)); roomID != RoomID {
			t.Errorf("message created but target room id is invalid")
		}
	}

	// case: not found resource
	{
		const (
			LoginUserID    = uint64(2)
			NotFoundRoomID = uint64(999)
		)

		qs := mocks.NewMockQueryService(mockCtrl)
		qs.EXPECT().
			FindRoomMessages(gomock.Any(), LoginUserID, gomock.Any()).
			Return(nil, chat.NewNotFoundError("")).
			Times(1)

		RESTHandler := &RESTHandler{chatQuery: qs}

		query := action.QueryRoomMessages{
			Before: action.TimestampNow(),
			Limit:  1,
		}
		req, err := newJSONRequest(echo.GET, URL, query)
		if err != nil {
			t.Fatal(err)
		}

		rec := httptest.NewRecorder()

		c := theEcho.NewContext(req, rec)
		c.Set(KeyLoggedInUserID, LoginUserID)
		c.SetParamNames("room_id")
		c.SetParamValues(fmt.Sprint(NotFoundRoomID))

		err = RESTHandler.GetRoomMessages(c)
		if err == nil {
			t.Error("given not found room id but no error")
		}
	}
}

func TestRESTGetUnreadRoomMessages(t *testing.T) {
	const URL = "/rooms/:room_id/messages/unread"

	t.Parallel()

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	// case: found with JSON
	{
		const (
			LoginUserID = uint64(2)
			RoomID      = uint64(3)
		)

		query := action.QueryUnreadRoomMessages{
			RoomID: RoomID,
			Limit:  1,
		}

		qs := mocks.NewMockQueryService(mockCtrl)
		qs.EXPECT().
			FindUnreadRoomMessages(gomock.Any(), LoginUserID, query).
			Return(&queried.EmptyUnreadRoomMessages, nil).
			Times(1)

		RESTHandler := &RESTHandler{chatQuery: qs}

		query.RoomID = 0 // remove RoomID from JSON
		req, err := newJSONRequest(echo.GET, URL, query)
		if err != nil {
			t.Fatal(err)
		}
		rec := httptest.NewRecorder()

		c := theEcho.NewContext(req, rec)
		c.Set(KeyLoggedInUserID, LoginUserID)
		c.SetParamNames("room_id")
		c.SetParamValues(fmt.Sprint(RoomID))

		err = RESTHandler.GetUnreadRoomMessages(c)
		if err != nil {
			t.Fatal(err)
		}
		if expect, got := http.StatusOK, rec.Code; expect != got {
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

	// case: found with query parameter
	{
		const (
			LoginUserID = uint64(2)
			RoomID      = uint64(3)
		)

		actionQuery := action.QueryUnreadRoomMessages{
			RoomID: RoomID,
			Limit:  1,
		}

		qs := mocks.NewMockQueryService(mockCtrl)
		qs.EXPECT().
			FindUnreadRoomMessages(gomock.Any(), LoginUserID, actionQuery).
			Return(&queried.EmptyUnreadRoomMessages, nil).
			Times(1)

		RESTHandler := &RESTHandler{chatQuery: qs}

		query := make(url.Values)
		query.Set("limit", fmt.Sprint(actionQuery.Limit))
		req := httptest.NewRequest(echo.GET, URL+"/?"+query.Encode(), nil)
		rec := httptest.NewRecorder()

		c := theEcho.NewContext(req, rec)
		c.Set(KeyLoggedInUserID, LoginUserID)
		c.SetParamNames("room_id")
		c.SetParamValues(fmt.Sprint(RoomID))

		err := RESTHandler.GetUnreadRoomMessages(c)
		if err != nil {
			t.Fatal(err)
		}
		if expect, got := http.StatusOK, rec.Code; expect != got {
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

	// case: room id is not found
	{
		const (
			NotFoundRoomID = uint64(999)
			LoginUserID    = uint64(2)
		)

		qs := mocks.NewMockQueryService(mockCtrl)
		qs.EXPECT().
			FindUnreadRoomMessages(gomock.Any(), LoginUserID, gomock.Any()).
			Return(nil, chat.NewNotFoundError("")).
			Times(1)

		RESTHandler := &RESTHandler{chatQuery: qs}

		query := action.QueryUnreadRoomMessages{
			Limit: 1,
		}

		req, err := newJSONRequest(echo.GET, "/rooms/:room_id/messages/unread", query)
		if err != nil {
			t.Fatal(err)
		}

		rec := httptest.NewRecorder()

		c := theEcho.NewContext(req, rec)
		c.Set(KeyLoggedInUserID, LoginUserID)
		c.SetParamNames("room_id")
		c.SetParamValues(fmt.Sprint(NotFoundRoomID))

		err = RESTHandler.GetUnreadRoomMessages(c)
		if err == nil {
			t.Error("given not found room id but no error")
		}
	}
}
