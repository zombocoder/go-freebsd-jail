//go:build freebsd

package vnet

import (
	"fmt"
	"strings"

	"github.com/zombocoder/go-freebsd-ifc/bridge"
	"github.com/zombocoder/go-freebsd-ifc/epair"
	ifc "github.com/zombocoder/go-freebsd-ifc/if"

	jailexec "github.com/zombocoder/go-freebsd-jail/exec"
	"github.com/zombocoder/go-freebsd-jail/jail"
)

// NetworkConfig describes the VNET networking setup for a jail.
type NetworkConfig struct {
	// JailName is the name of the target jail.
	JailName string

	// Bridge is the name of an existing bridge interface to use.
	Bridge string

	// CreateBridge creates the bridge if true.
	CreateBridge bool

	// EpairHost is the host-side epair name. If empty, auto-created.
	EpairHost string

	// EpairJail is the jail-side epair name. If empty, auto-created.
	EpairJail string

	// CreateEpair creates a new epair if true (EpairHost/EpairJail are set on return).
	CreateEpair bool

	// JailInterfaceName is the desired interface name inside the jail (e.g. "eth0").
	// If empty, the epair name is kept.
	JailInterfaceName string

	// HostAddress is the IP address for the host-side bridge (e.g. "10.0.0.1/24").
	// Optional — set if you need the host to be on the same subnet.
	HostAddress string

	// JailAddress is the IP address to assign inside the jail (e.g. "10.0.0.2/24").
	JailAddress string

	// Gateway is the default gateway to set inside the jail (e.g. "10.0.0.1").
	Gateway string
}

// Setup configures VNET networking for a jail.
//
// Sequence:
//  1. Create epair (if CreateEpair)
//  2. Create bridge (if CreateBridge)
//  3. Add host-side epair to bridge
//  4. Bring up host-side epair
//  5. Move jail-side epair into jail via jail.Update
//  6. Configure IP and route inside jail via exec.Shell
func Setup(cfg NetworkConfig) error {
	if cfg.JailName == "" {
		return &jail.OperationError{Op: "vnet.setup", Jail: cfg.JailName, Err: jail.ErrInvalidArgument}
	}

	// Step 1: Create epair if requested
	if cfg.CreateEpair {
		pair, err := epair.Create()
		if err != nil {
			return &jail.OperationError{Op: "vnet.setup", Jail: cfg.JailName, Err: fmt.Errorf("create epair: %w", err)}
		}
		cfg.EpairHost = pair.A
		cfg.EpairJail = pair.B
	}

	if cfg.EpairHost == "" || cfg.EpairJail == "" {
		return &jail.OperationError{Op: "vnet.setup", Jail: cfg.JailName, Err: fmt.Errorf("%w: epair names must be set", jail.ErrInvalidArgument)}
	}

	// Step 2: Create bridge if requested
	if cfg.CreateBridge {
		name, err := bridge.Create()
		if err != nil {
			return &jail.OperationError{Op: "vnet.setup", Jail: cfg.JailName, Err: fmt.Errorf("create bridge: %w", err)}
		}
		cfg.Bridge = name
		if err := bridge.Up(cfg.Bridge, true); err != nil {
			return &jail.OperationError{Op: "vnet.setup", Jail: cfg.JailName, Err: fmt.Errorf("bridge up: %w", err)}
		}
	}

	// Step 3: Add host-side epair to bridge (idempotent)
	if cfg.Bridge != "" {
		if err := bridge.AddMember(cfg.Bridge, cfg.EpairHost); err != nil {
			return &jail.OperationError{Op: "vnet.setup", Jail: cfg.JailName, Err: fmt.Errorf("bridge add member: %w", err)}
		}
	}

	// Step 4: Bring up host-side epair
	if err := ifc.SetUp(cfg.EpairHost, true); err != nil {
		return &jail.OperationError{Op: "vnet.setup", Jail: cfg.JailName, Err: fmt.Errorf("epair up: %w", err)}
	}

	// Step 5: Move jail-side epair into jail
	if err := jail.Update(cfg.JailName, jail.Config{
		VNETInterface: []string{cfg.EpairJail},
	}); err != nil {
		return &jail.OperationError{Op: "vnet.setup", Jail: cfg.JailName, Err: fmt.Errorf("move interface to jail: %w", err)}
	}

	// Step 6: Configure networking inside the jail via exec
	ifName := cfg.EpairJail
	if cfg.JailInterfaceName != "" && cfg.JailInterfaceName != cfg.EpairJail {
		cmd := fmt.Sprintf("ifconfig %s name %s", cfg.EpairJail, cfg.JailInterfaceName)
		if err := jailexec.Shell(cfg.JailName, cmd); err != nil {
			return &jail.OperationError{Op: "vnet.setup", Jail: cfg.JailName, Err: fmt.Errorf("rename interface: %w", err)}
		}
		ifName = cfg.JailInterfaceName
	}

	// Bring up jail-side interface
	if err := jailexec.Shell(cfg.JailName, fmt.Sprintf("ifconfig %s up", ifName)); err != nil {
		return &jail.OperationError{Op: "vnet.setup", Jail: cfg.JailName, Err: fmt.Errorf("interface up: %w", err)}
	}

	// Assign IP address
	if cfg.JailAddress != "" {
		addr := cfg.JailAddress
		// Determine if IPv4 or IPv6
		if strings.Contains(addr, ":") {
			cmd := fmt.Sprintf("ifconfig %s inet6 %s", ifName, addr)
			if err := jailexec.Shell(cfg.JailName, cmd); err != nil {
				return &jail.OperationError{Op: "vnet.setup", Jail: cfg.JailName, Err: fmt.Errorf("assign ip6: %w", err)}
			}
		} else {
			cmd := fmt.Sprintf("ifconfig %s inet %s", ifName, addr)
			if err := jailexec.Shell(cfg.JailName, cmd); err != nil {
				return &jail.OperationError{Op: "vnet.setup", Jail: cfg.JailName, Err: fmt.Errorf("assign ip4: %w", err)}
			}
		}
	}

	// Set default route
	if cfg.Gateway != "" {
		if strings.Contains(cfg.Gateway, ":") {
			cmd := fmt.Sprintf("route add -6 default %s", cfg.Gateway)
			if err := jailexec.Shell(cfg.JailName, cmd); err != nil {
				return &jail.OperationError{Op: "vnet.setup", Jail: cfg.JailName, Err: fmt.Errorf("add default route6: %w", err)}
			}
		} else {
			cmd := fmt.Sprintf("route add default %s", cfg.Gateway)
			if err := jailexec.Shell(cfg.JailName, cmd); err != nil {
				return &jail.OperationError{Op: "vnet.setup", Jail: cfg.JailName, Err: fmt.Errorf("add default route: %w", err)}
			}
		}
	}

	return nil
}

// Teardown removes VNET networking for a jail.
//
// Idempotent: missing interfaces/bridges are silently ignored.
func Teardown(cfg NetworkConfig) error {
	if cfg.JailName == "" {
		return &jail.OperationError{Op: "vnet.teardown", Jail: cfg.JailName, Err: jail.ErrInvalidArgument}
	}

	// Remove host-side epair from bridge (idempotent)
	if cfg.Bridge != "" && cfg.EpairHost != "" {
		_ = bridge.DelMember(cfg.Bridge, cfg.EpairHost)
	}

	// Destroy epair (destroys both sides, idempotent)
	if cfg.EpairHost != "" {
		_ = epair.Destroy(cfg.EpairHost)
	}

	// Destroy bridge if we created it
	if cfg.CreateBridge && cfg.Bridge != "" {
		_ = bridge.Destroy(cfg.Bridge)
	}

	return nil
}
