package jail

import "github.com/zombocoder/go-freebsd-jail/internal/jailerr"

// Sentinel errors for jail operations.
// Re-exported from internal/jailerr to provide a public API.
var (
	ErrPermission      = jailerr.ErrPermission
	ErrNotFound        = jailerr.ErrNotFound
	ErrExists          = jailerr.ErrExists
	ErrInvalidArgument = jailerr.ErrInvalidArgument
	ErrBusy            = jailerr.ErrBusy
	ErrNotSupported    = jailerr.ErrNotSupported
	ErrNoMemory        = jailerr.ErrNoMemory
)

// OperationError wraps an error with context about the jail operation that failed.
type OperationError = jailerr.OperationError

// ValidationError represents an input validation failure.
type ValidationError = jailerr.ValidationError
