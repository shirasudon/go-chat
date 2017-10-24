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
	cause := fmt.Errorf(msgFormat, args...)
	return &InfraError{Cause: cause}
}

func (err InfraError) Error() string {
	return fmt.Sprintf("infra error: %v", err.Cause.Error())
}

// ErrInternalError is the proxy for the internal error
// which should not be shown for the client.
var ErrInternalError = errors.New("Internal error")
