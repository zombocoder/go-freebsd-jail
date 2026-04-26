package unit

import (
	"errors"
	"testing"

	"github.com/zombocoder/go-freebsd-jail/jail"
	"github.com/zombocoder/go-freebsd-jail/internal/jailerr"
)

func TestErrorsIs(t *testing.T) {
	// Verify that jail.Err* matches jailerr.Err* (they are the same pointers)
	if !errors.Is(jail.ErrNotFound, jailerr.ErrNotFound) {
		t.Error("jail.ErrNotFound should match jailerr.ErrNotFound")
	}
	if !errors.Is(jail.ErrPermission, jailerr.ErrPermission) {
		t.Error("jail.ErrPermission should match jailerr.ErrPermission")
	}
}

func TestOperationError_Unwrap(t *testing.T) {
	err := &jail.OperationError{
		Op:   "create",
		Jail: "web",
		Err:  jail.ErrPermission,
	}

	if !errors.Is(err, jail.ErrPermission) {
		t.Error("OperationError should unwrap to ErrPermission")
	}

	msg := err.Error()
	if msg == "" {
		t.Error("OperationError.Error() should not be empty")
	}
}

func TestOperationError_WithParam(t *testing.T) {
	err := &jail.OperationError{
		Op:    "create",
		Jail:  "web",
		Param: "path",
		Err:   jail.ErrInvalidArgument,
	}

	if !errors.Is(err, jail.ErrInvalidArgument) {
		t.Error("should unwrap to ErrInvalidArgument")
	}
}

func TestValidationError_Unwrap(t *testing.T) {
	err := &jail.ValidationError{
		Field: "name",
		Value: "",
		Err:   jail.ErrInvalidArgument,
	}

	if !errors.Is(err, jail.ErrInvalidArgument) {
		t.Error("ValidationError should unwrap to ErrInvalidArgument")
	}

	var ve *jail.ValidationError
	if !errors.As(err, &ve) {
		t.Error("errors.As should work with *ValidationError")
	}
	if ve.Field != "name" {
		t.Errorf("expected field %q, got %q", "name", ve.Field)
	}
}
