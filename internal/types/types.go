// Package types defines shared types used across the jail library.
// This package has no imports from other internal packages to avoid import cycles.
package types

// JID represents a jail identifier.
type JID int

// State represents the current state of a jail.
type State string

const (
	StateActive State = "active"
	StateDying  State = "dying"
)

// Config defines parameters for creating or updating a jail.
type Config struct {
	Name          string
	Path          string
	Hostname      string
	Persist       bool
	IP4           []string
	IP6           []string
	VNET          bool
	VNETInterface []string
	MountDevFS    bool
	DevFSRuleset  *int
	Allow         AllowConfig
	Limits        LimitConfig
	RawParams     map[string]string
}

// AllowConfig controls what operations are permitted inside the jail.
type AllowConfig struct {
	RawSockets  bool
	Mount       bool
	MountDevFS  bool
	MountNullFS bool
	MountProcFS bool
	MountTmpFS  bool
	SysVIPC     bool
	SetHostname bool
	Chflags     bool
	SocketAF    bool
}

// LimitConfig controls resource limits for the jail.
type LimitConfig struct {
	ChildrenMax *int
	CPUSetsize  *int
}

// Info contains information about a running jail.
type Info struct {
	JID      JID
	Name     string
	Path     string
	Hostname string
	State    State
	IP4      []string
	IP6      []string
	VNET     bool
	Params   map[string]string
}
