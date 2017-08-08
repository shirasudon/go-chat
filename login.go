package chat

import (
	"encoding/gob"
	"net/http"

	"github.com/ipfans/echo-session"
	"github.com/labstack/echo"
	"github.com/shirasudon/go-chat/entity"
)

func init() {
	// register LoginState which is requirements
	// to use echo-session and backed-end gorilla/sessions.
	gob.Register(&LoginState{})
}

type UserForm struct {
	Email      string `json:"email" form:"email" query:"name"`
	Password   string `json:"password" form:"password" query:"password"`
	RememberMe bool   `json:"rememberMe" form:"rememberMe" query:"rememberMe"`
}

type LoginState struct {
	LoggedIn   bool   `json:"logged_in"`
	RememberMe bool   `json:"rememberMe"`
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
	userRepo entity.UserRepository
	store    session.Store
}

func NewLoginHandler(uRepo entity.UserRepository, secretKeyPairs ...[]byte) *LoginHandler {
	if len(secretKeyPairs) == 0 {
		secretKeyPairs = [][]byte{
			[]byte("sercret-key"),
		}
	}
	store := session.NewCookieStore(secretKeyPairs...)
	store.Options(DefaultOptions)

	return &LoginHandler{
		userRepo: uRepo,
		store:    store,
	}
}

func (lh *LoginHandler) Login(c echo.Context) error {
	u := new(UserForm)
	if err := c.Bind(u); err != nil {
		return err
	}

	user, err := lh.userRepo.Get(u.Email, u.Password)
	if err != nil {
		return c.JSON(http.StatusOK, LoginState{ErrorMsg: err.Error()})
	}

	loginState := LoginState{LoggedIn: true, UserID: user.ID, RememberMe: u.RememberMe}

	sess := session.Default(c)
	sess.Set(KeyLoginState, loginState)
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
	if _, ok := sess.Get(KeyLoginState).(LoginState); !ok {
		return c.JSON(http.StatusOK, LoginState{ErrorMsg: "you are not logged in"})
	}
	sess.Delete(KeyLoginState)
	if err := sess.Save(); err != nil {
		return err
	}
	return c.JSON(http.StatusOK, LoginState{LoggedIn: false})
}

func (lh *LoginHandler) GetLoginState(c echo.Context) error {
	sess := session.Default(c)
	loginState, ok := sess.Get(KeyLoginState).(LoginState)
	if !ok {
		return c.JSON(http.StatusOK, LoginState{LoggedIn: false, ErrorMsg: "you are not logged in"})
	}
	return c.JSON(http.StatusOK, loginState)
}

func (lh *LoginHandler) IsLoggedInRequest(c echo.Context) bool {
	loginState, ok := session.Default(c).Get(KeyLoginState).(LoginState)
	return ok && loginState.LoggedIn
}

// Middleware returns echo.MiddlewareFunc.
// it should be registered for echo.Server to use this LoginHandler.
func (lh *LoginHandler) Middleware() echo.MiddlewareFunc {
	return session.Sessions(KeySessionID, lh.store)
}

// Filter is a middleware which filters unauthorized request.
func (lh *LoginHandler) Filter() echo.MiddlewareFunc {
	return func(handlerFunc echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if lh.IsLoggedInRequest(c) {
				return handlerFunc(c)
			}
			return c.JSON(http.StatusForbidden, struct {
				ErrorMsg string `json:"error"`
			}{"you are not logged in"})
		}
	}
}
