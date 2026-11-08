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
	Name     string
	Path     string
	Hostname string
	Persist  bool
	IP4      []string
	IP6      []string
	VNET     bool

	// VNETInterface lists host interfaces to move into the jail's vnet
	// (requires VNET). Like jail(8)'s vnet.interface, but applied with the
	// SIOCSIFVNET ioctl after the jail is created, since the kernel has no
	// such jail parameter.
	VNETInterface []string

	// MountDevFS mounts a devfs on <Path>/dev before the jail is created,
	// like jail(8)'s mount.devfs, applying DevFSRuleset (default 4,
	// devfsrules_jail). Remove unmounts it once no other jail uses Path.
	// <Path>/dev must be a real directory, not a symlink, and Path must
	// not resolve to "/".
	MountDevFS bool

	// DevFSRuleset sets the devfs_ruleset jail parameter and the ruleset
	// applied to the devfs mounted by MountDevFS.
	DevFSRuleset *int
	Allow        AllowConfig
	Limits       LimitConfig
	RawParams    map[string]string
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
