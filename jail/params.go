package jail

import "github.com/zombocoder/go-freebsd-jail/internal/types"

// State represents the current state of a jail.
type State = types.State

const (
	// StateActive indicates the jail is running.
	StateActive = types.StateActive

	// StateDying indicates the jail is being removed.
	StateDying = types.StateDying
)

// Info contains information about a running jail.
type Info = types.Info
