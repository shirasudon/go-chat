package chat

import (
	"errors"
	"testing"
)

func TestInfraError(t *testing.T) {
	for _, err := range []error{
		NewInfraError("error test"),
		NewInfraError("error test %v", "message"),
		NewInfraError(""),
	} {
		if err == nil {
			t.Fatal("can not create new infra error")
		}
		if msg := err.Error(); len(msg) == 0 {
			t.Error("new infra error shows no message")
		}
	}
}

func TestNotFoundError(t *testing.T) {
	for _, err := range []error{
		NewNotFoundError("error test"),
		NewNotFoundError("error test %v", "message"),
		NewNotFoundError(""),
	} {
		if err == nil {
			t.Fatal("can not create new not found error")
		}
		if msg := err.Error(); len(msg) == 0 {
			t.Error("new not found error shows no message")
		}
		if _, ok := err.(*NotFoundError); !ok {
			t.Errorf("invalid error type, %T", err)
		}
		if _, ok := err.(*InfraError); ok {
			t.Errorf("passing other type assertion")
		}
	}
}

func TestIsNotFoundError(t *testing.T) {
	for _, tcase := range []struct {
		Err             error
		IsNotFoundError bool
	}{
		{NewNotFoundError(""), true},
		{NotFoundError{}, true},
		{NewInfraError(""), false},
		{errors.New(""), false},
		{nil, false},
	} {
		if IsNotFoundError(tcase.Err) != tcase.IsNotFoundError {
			t.Errorf("%T is detected as NotFoundError", tcase.Err)
		}
	}
}
