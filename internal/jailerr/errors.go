// Package jailerr defines sentinel errors used across the jail library.
// This package has no imports from other internal packages to avoid import cycles.
package jailerr

import (
	"errors"
	"fmt"
)

// Sentinel errors for jail operations.
var (
	ErrPermission      = errors.New("permission denied")
	ErrNotFound        = errors.New("jail not found")
	ErrExists          = errors.New("jail already exists")
	ErrInvalidArgument = errors.New("invalid argument")
	ErrBusy            = errors.New("resource busy")
	ErrNotSupported    = errors.New("operation not supported")
	ErrNoMemory        = errors.New("out of memory")
)

// OperationError wraps an error with context about the jail operation that failed.
type OperationError struct {
	Op    string
	Jail  string
	Param string
	Err   error
}

func (e *OperationError) Error() string {
	if e.Param != "" {
		return fmt.Sprintf("jail %s %s: param %s: %v", e.Op, e.Jail, e.Param, e.Err)
	}
	if e.Jail != "" {
		return fmt.Sprintf("jail %s %s: %v", e.Op, e.Jail, e.Err)
	}
	return fmt.Sprintf("jail %s: %v", e.Op, e.Err)
}

func (e *OperationError) Unwrap() error {
	return e.Err
}

// ValidationError represents an input validation failure.
type ValidationError struct {
	Field string
	Value string
	Err   error
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("invalid %s %q: %v", e.Field, e.Value, e.Err)
}

func (e *ValidationError) Unwrap() error {
	return e.Err
}
