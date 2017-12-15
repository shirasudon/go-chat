package chat

import (
	"errors"
	"fmt"
)

// TODO distinguish errors from infrastructures and domain.
// For example, infrastructure errors contains NotFound,
// TxFailed and Operaion Failed etc.

// InfraError represents error caused on the
// infrastructure layer.
// It should be not shown directly for the client side.
//
// It implements error interface.
type InfraError struct {
	Cause error
}

// NewInfraError create new InfraError with same syntax as fmt.Errorf().
func NewInfraError(msgFormat string, args ...interface{}) *InfraError {
	return &InfraError{Cause: fmt.Errorf(msgFormat, args...)}
}

func (err InfraError) Error() string {
	return fmt.Sprintf("infra error: %v", err.Cause.Error())
}

// ErrInternalError is the proxy for the internal error
// which should not be shown for the client.
var ErrInternalError = errors.New("Internal error")

// NotFoundError represents error caused on the
// infrastructure layer but it is specified to that
// the request data is not found.
// It can be shown directly for the client side.
//
// It implements error interface.
type NotFoundError InfraError

// NewInfraError create new InfraError with same syntax as fmt.Errorf().
func NewNotFoundError(msgFormat string, args ...interface{}) *NotFoundError {
	return &NotFoundError{Cause: fmt.Errorf(msgFormat, args...)}
}

func (err NotFoundError) Error() string {
	return fmt.Sprintf("not found error: %v", err.Cause.Error())
}

// It returns true when the type of given err is *NotFoundError or NotFoundError,
// otherwise false.
func IsNotFoundError(err error) bool {
	switch err.(type) {
	case NotFoundError, *NotFoundError:
		return true
	default:
		return false
	}
}
