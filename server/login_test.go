package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/ipfans/echo-session"
	"github.com/labstack/echo"

	"github.com/shirasudon/go-chat/chat"
	"github.com/shirasudon/go-chat/chat/queried"
	"github.com/shirasudon/go-chat/internal/mocks"
)

func NewMockLoginHandler(ctrl *gomock.Controller) (*LoginHandler, *mocks.MockLoginService) {
	ls := mocks.NewMockLoginService(ctrl)
	return NewLoginHandler(ls), ls
}

const (
	CorrectName     = "user"
	CorrectPassword = "password"
)

var AuthUser = queried.AuthUser{
	ID:       2,
	Name:     CorrectName,
	Password: CorrectPassword,
}

func TestLogin(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	var AuthUser = queried.AuthUser{
		ID:       2,
		Name:     "user",
		Password: "password",
	}

	// correct user login
	loginHandler, service := NewMockLoginHandler(ctrl)
	service.EXPECT().Login(gomock.Any(), CorrectName, CorrectPassword).
		Return(&AuthUser, nil).Times(1)

	c, err := doLogin(loginHandler, CorrectName, CorrectPassword, true)
	if err != nil {
		t.Fatalf("can not login: %v", err)
	}

	// check session has LoginState.
	sess := session.Default(c)
	if ls, ok := sess.Get(KeyLoginState).(*LoginState); !ok || !ls.LoggedIn {
		t.Errorf("session does not have any LoginState after login")
	}

	// check response json
	loginState, err := loginStateFromResponse(c)
	if err != nil {
		t.Fatal(err)
	}
	if loginState.UserID != AuthUser.ID {
		t.Errorf("different logged in user ID, expect: %v, got: %v", AuthUser.ID, loginState.UserID)
	}
	if !loginState.LoggedIn {
		t.Error("can not logged in")
	}
	if !loginState.RememberMe {
		t.Error("login with RememberMe but not set")
	}
	if msg := loginState.ErrorMsg; len(msg) > 0 {
		t.Errorf("login succeeded but got error: %v", msg)
	}
}

func TestLoginFail(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	loginHandler, service := NewMockLoginHandler(ctrl)
	service.EXPECT().Login(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(nil, chat.NewNotFoundError("not found")).AnyTimes()

	// wrong user login
	for _, testcase := range []struct {
		Name     string
		Password string
	}{
		{"wrong", CorrectPassword},
		{CorrectName, "wrong"},
		{"wrong", "wrong"},
	} {
		c, err := doLogin(loginHandler, testcase.Name, testcase.Password, false)
		if err != nil {
			t.Fatalf("got error: login with email: %v password: %v, err: %v", testcase.Name, testcase.Password, err)
		}

		// check session has LoginState.
		sess := session.Default(c)
		if _, ok := sess.Get(KeyLoginState).(*LoginState); ok {
			t.Errorf("session has any LoginState after login failed with email: %v, password: %v", testcase.Name, testcase.Password)
		}

		// check response json
		loginState, err := loginStateFromResponse(c)
		if err != nil {
			t.Fatal(err)
		}
		if loginState.LoggedIn {
			t.Errorf("LoggedIn is true after login failed with email: %v, password: %v", testcase.Name, testcase.Password)
		}
		if loginState.RememberMe {
			t.Error("login without RememberMe but set")
		}
		if msg := loginState.ErrorMsg; len(msg) == 0 {
			t.Errorf("missing ErrorMsg after login failed with email: %v, password: %v", testcase.Name, testcase.Password)
		}
	}
}

func doLogin(lh *LoginHandler, name, password string, rememberMe bool) (echo.Context, error) {
	// POST form with email and password
	f := make(url.Values)
	f.Set("name", name)
	f.Set("password", password)
	f.Set("remember_me", fmt.Sprint(rememberMe))
	req := httptest.NewRequest(echo.POST, "/login", strings.NewReader(f.Encode()))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec := httptest.NewRecorder()

	c := theEcho.NewContext(req, rec)
	return c, withSession(lh, lh.Login, c)
}

// with session set loggedin state before invoke handler.
func withSession(loginHandler *LoginHandler, hf echo.HandlerFunc, c echo.Context) error {
	return loginHandler.Middleware()(hf)(c)
}

func loginStateFromResponse(c echo.Context) (LoginState, error) {
	loginState := LoginState{}
	rec := c.Response().Writer.(*httptest.ResponseRecorder)
	return loginState, json.Unmarshal(rec.Body.Bytes(), &loginState)
}

func TestLogout(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	loginHandler, service := NewMockLoginHandler(ctrl)
	service.EXPECT().Login(gomock.Any(), CorrectName, CorrectPassword).Return(&AuthUser, nil).Times(1)
	service.EXPECT().Logout(gomock.Any(), AuthUser.ID).Times(1)

	// firstly we try to logout without login.
	// first call not passing service.Logout()
	c, err := doLogout(loginHandler, nil)
	if err != nil {
		t.Fatal(err)
	}

	// check logout response
	loginState, err := loginStateFromResponse(c)
	if err != nil {
		t.Fatal(err)
	}
	if loginState.LoggedIn {
		t.Errorf("logout response without login expects LoggedIn = %v but %v", false, loginState.LoggedIn)
	}
	if len(loginState.ErrorMsg) == 0 {
		t.Errorf("logout response without login expects some ErrorMsg but no message")
	}

	// secondary, we try to logout after logged in.
	c, _ = doLogin(loginHandler, CorrectName, CorrectPassword, false)

	c, err = doLogout(loginHandler, c.Response().Header()["Set-Cookie"])
	if err != nil {
		t.Fatal(err)
	}

	// check session has no loginState
	sess := session.Default(c)
	if _, ok := sess.Get(KeyLoginState).(*LoginState); ok {
		t.Errorf("session have LoginState after logout")
	}

	// check logout response
	loginState, err = loginStateFromResponse(c)
	if err != nil {
		t.Fatal(err)
	}
	if loginState.LoggedIn {
		t.Errorf("logout response after login expects LoggedIn = %v but %v", false, loginState.LoggedIn)
	}
	if msg := loginState.ErrorMsg; len(msg) > 0 {
		t.Errorf("logout response after login expects no erorr message but message=%v", msg)
	}
}

func doLogout(lh *LoginHandler, cookie []string) (echo.Context, error) {
	req := httptest.NewRequest(echo.POST, "/logout", nil)
	req.Header["Cookie"] = cookie
	rec := httptest.NewRecorder()
	c := theEcho.NewContext(req, rec)
	return c, withSession(lh, lh.Logout, c)
}

func TestGetLoginState(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// before logged-in, it returns loginState with loggedin=false
	loginHandler, service := NewMockLoginHandler(ctrl)
	service.EXPECT().Login(gomock.Any(), CorrectName, CorrectPassword).Return(&AuthUser, nil).Times(1)

	c, err := doGetLoginState(loginHandler, nil)
	if err != nil {
		t.Fatal(err)
	}

	// check response json
	loginState, err := loginStateFromResponse(c)
	if err != nil {
		t.Fatal(err)
	}
	if loginState.LoggedIn {
		t.Error("not loggedin but loginState.LoggedIn = true")
	}
	if msg := loginState.ErrorMsg; len(msg) == 0 {
		t.Errorf("not loggedin but no error")
	}

	// after logged-in, it returns loginState with LoggedIn = true.
	c, _ = doLogin(loginHandler, CorrectName, CorrectPassword, false)

	c, err = doGetLoginState(loginHandler, c.Response().Header()["Set-Cookie"])
	if err != nil {
		t.Fatal(err)
	}

	// check response json
	loginState, err = loginStateFromResponse(c)
	if err != nil {
		t.Fatal(err)
	}
	if !loginState.LoggedIn {
		t.Error("login succeeded but can not get logged in state")
	}
	if msg := loginState.ErrorMsg; len(msg) > 0 {
		t.Errorf("login succeeded but got error: %v", msg)
	}
}

func doGetLoginState(lh *LoginHandler, cookie []string) (echo.Context, error) {
	req := httptest.NewRequest(echo.GET, "/login", nil)
	req.Header["Cookie"] = cookie
	rec := httptest.NewRecorder()
	c := theEcho.NewContext(req, rec)
	return c, withSession(lh, lh.GetLoginState, c)
}

func TestIsLoggedinRequest(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	loginHandler, service := NewMockLoginHandler(ctrl)
	service.EXPECT().Login(gomock.Any(), CorrectName, CorrectPassword).Return(&AuthUser, nil).Times(1)

	req := httptest.NewRequest(echo.GET, "/", nil)
	rec := httptest.NewRecorder()
	c := theEcho.NewContext(req, rec)
	if loginHandler.IsLoggedInRequest(c) {
		t.Error("without logged-in request but IsLoggedInRequest returns true")
	}

	c, _ = doLogin(loginHandler, CorrectName, CorrectPassword, false)
	if !loginHandler.IsLoggedInRequest(c) {
		t.Error("with logged-in request but IsLoggedInRequest returns false")
	}
}

func TestLoginFilter(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	var (
		ErrHandlerPassed = fmt.Errorf("handler passed")
		ErrHandler       = func(c echo.Context) error {
			return ErrHandlerPassed
		}
	)

	// case1: not logged in
	{
		loginHandler, _ := NewMockLoginHandler(ctrl)
		filteredHandler := loginHandler.Filter()(ErrHandler)

		req := httptest.NewRequest(echo.GET, "/", nil)
		rec := httptest.NewRecorder()
		c := theEcho.NewContext(req, rec)
		err := filteredHandler(c)
		if err == nil {
			t.Fatal("without logged-in request but not filtered")
		}
		herr, ok := err.(*echo.HTTPError)
		if !ok {
			t.Fatal("filtered error is not a HTTPError")
		}
		if herr.Code != http.StatusForbidden {
			t.Errorf("differenct http status, expect: %v, got: %v", http.StatusText(http.StatusForbidden), http.StatusText(herr.Code))
		}
		if len(herr.Error()) == 0 {
			t.Error("empty error message")
		}
	}

	// case2: with logged in
	{
		loginHandler, service := NewMockLoginHandler(ctrl)
		service.EXPECT().Login(gomock.Any(), CorrectName, CorrectPassword).Return(&AuthUser, nil).Times(1)

		filteredHandler := loginHandler.Filter()(ErrHandler)

		c, _ := doLogin(loginHandler, CorrectName, CorrectPassword, false)
		err := filteredHandler(c)
		if err != ErrHandlerPassed {
			t.Fatalf("with login request, but handler is not executed: %v", err)
		}
	}
}
