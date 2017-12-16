package chat

import (
	"encoding/gob"
	"net/http"

	"github.com/ipfans/echo-session"
	"github.com/labstack/echo"

	"github.com/shirasudon/go-chat/chat"
)

func init() {
	// register LoginState which is requirements
	// to use echo-session and backed-end gorilla/sessions.
	gob.Register(&LoginState{})
}

type UserForm struct {
	Name       string `json:"name" form:"name" query:"name"`
	Password   string `json:"password" form:"password" query:"password"`
	RememberMe bool   `json:"remember_me" form:"remember_me" query:"remember_me"`
}

type LoginState struct {
	LoggedIn   bool   `json:"logged_in"`
	RememberMe bool   `json:"remember_me"`
	UserID     uint64 `json:"user_id"`
	ErrorMsg   string `json:"error,omitempty"`
}

const (
	KeySessionID = "SESSION-ID"

	// key for session value which is user loggedin state.
	KeyLoginState = "LOGIN-STATE"

	// seconds in 365 days, where 86400 is a seconds in 1 day
	SecondsInYear = 86400 * 365
)

var DefaultOptions = session.Options{
	HttpOnly: true,
}

// LoginHandler handles login requests.
// it holds logged-in users, so that each request can reference
// any logged-in user.
type LoginHandler struct {
	users chat.UserQueryer
	store session.Store
}

func NewLoginHandler(users chat.UserQueryer, secretKeyPairs ...[]byte) *LoginHandler {
	if len(secretKeyPairs) == 0 {
		secretKeyPairs = [][]byte{
			[]byte("secret-key"),
		}
	}
	store := session.NewCookieStore(secretKeyPairs...)
	store.Options(DefaultOptions)

	return &LoginHandler{
		users: users,
		store: store,
	}
}

func (lh *LoginHandler) Login(c echo.Context) error {
	u := new(UserForm)
	if err := c.Bind(u); err != nil {
		return err
	}

	user, err := lh.users.FindByNameAndPassword(c.Request().Context(), u.Name, u.Password)
	if err != nil {
		return c.JSON(http.StatusOK, LoginState{ErrorMsg: err.Error()})
	}

	loginState := LoginState{LoggedIn: true, UserID: user.ID, RememberMe: u.RememberMe}

	sess := session.Default(c)
	sess.Set(KeyLoginState, &loginState)
	if loginState.RememberMe {
		newOpt := DefaultOptions
		newOpt.MaxAge = SecondsInYear
		sess.Options(newOpt)
	}
	if err := sess.Save(); err != nil {
		return err
	}

	return c.JSON(http.StatusOK, loginState)
}

func (lh *LoginHandler) Logout(c echo.Context) error {
	sess := session.Default(c)
	if _, ok := sess.Get(KeyLoginState).(*LoginState); !ok {
		return c.JSON(http.StatusOK, LoginState{ErrorMsg: "you are not logged in"})
	}
	sess.Delete(KeyLoginState)
	if err := sess.Save(); err != nil {
		return err
	}
	return c.JSON(http.StatusOK, LoginState{LoggedIn: false})
}

func (lh *LoginHandler) GetLoginState(c echo.Context) error {
	loginState, ok := lh.Session(c)
	if !ok {
		return c.JSON(http.StatusOK, LoginState{LoggedIn: false, ErrorMsg: "you are not logged in"})
	}
	return c.JSON(http.StatusOK, loginState)
}

func (lh *LoginHandler) IsLoggedInRequest(c echo.Context) bool {
	loginState, ok := lh.Session(c)
	return ok && loginState.LoggedIn
}

// it returns loginState as session state.
// the second returned value is true when LoginState exists.
func (lh *LoginHandler) Session(c echo.Context) (*LoginState, bool) {
	sess := session.Default(c)
	if sess == nil {
		return nil, false
	}
	loginState, ok := sess.Get(KeyLoginState).(*LoginState)
	return loginState, ok
}

// Middleware returns echo.MiddlewareFunc.
// it should be registered for echo.Server to use this LoginHandler.
func (lh *LoginHandler) Middleware() echo.MiddlewareFunc {
	return session.Sessions(KeySessionID, lh.store)
}

// KeyLoggedInUserID is the key for the logged-in user id in the echo.Context.
// It is set when user is logged-in through LoginHandler.
const KeyLoggedInUserID = "SESSION-USER-ID"

// Filter is a middleware which filters unauthenticated request.
//
// it sets logged-in user's id for echo.Context using KeyLoggedInUserID
// when the request is authenticated.
func (lh *LoginHandler) Filter() echo.MiddlewareFunc {
	return func(handlerFunc echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if loginState, ok := lh.Session(c); ok && loginState.LoggedIn {
				c.Set(KeyLoggedInUserID, loginState.UserID)
				return handlerFunc(c)
			}
			// not logged-in
			return NewHTTPError(http.StatusForbidden, "require login firstly")
		}
	}
}

// get logged in user id which is valid after LoginHandler.Filter.
// the second returned value is false if logged in
// user id is not found.
func LoggedInUserID(c echo.Context) (uint64, bool) {
	userID, ok := c.Get(KeyLoggedInUserID).(uint64)
	return userID, ok
}
