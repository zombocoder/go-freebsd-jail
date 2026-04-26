package jail

import "github.com/zombocoder/go-freebsd-jail/internal/types"

// JID represents a jail identifier.
type JID = types.JID

// Config defines parameters for creating or updating a jail.
type Config = types.Config

// AllowConfig controls what operations are permitted inside the jail.
type AllowConfig = types.AllowConfig

// LimitConfig controls resource limits for the jail.
type LimitConfig = types.LimitConfig
