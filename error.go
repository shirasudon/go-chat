package chat

import (
	"github.com/labstack/echo"
)

// Wrapper function for the echo.HTTPError.
// Because echo.DefaultHTTPHandler handles only string message in
// the echo.HTTPError, This converts error types to string.
func NewHTTPError(statusCode int, msg ...interface{}) *echo.HTTPError {
	if len(msg) > 0 {
		var strMsg = ""
		switch m := msg[0].(type) {
		case string:
			strMsg = m
		case error:
			strMsg = m.Error()
		default:
			panic("NewHTTPError: support only string or error")
		}
		return echo.NewHTTPError(statusCode, strMsg)
	}
	return echo.NewHTTPError(statusCode)
}
