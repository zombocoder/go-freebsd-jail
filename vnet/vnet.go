//go:build freebsd

package vnet

import (
	"errors"
	"fmt"
	"net"
	"strings"

	"github.com/zombocoder/go-freebsd-ifc/bridge"
	"github.com/zombocoder/go-freebsd-ifc/epair"
	ifc "github.com/zombocoder/go-freebsd-ifc/if"
	"github.com/zombocoder/go-freebsd-ifc/ip"

	jailexec "github.com/zombocoder/go-freebsd-jail/exec"
	"github.com/zombocoder/go-freebsd-jail/jail"
)

// NetworkConfig describes the VNET networking setup for a jail.
type NetworkConfig struct {
	// JailName is the name of the target jail.
	JailName string

	// Bridge is the name of an existing bridge interface to use.
	// Set on return when CreateBridge is true.
	Bridge string

	// CreateBridge creates the bridge if true.
	CreateBridge bool

	// EpairHost is the host-side epair name. Set on return when CreateEpair is true.
	EpairHost string

	// EpairJail is the jail-side epair name. Set on return when CreateEpair is true.
	EpairJail string

	// CreateEpair creates a new epair if true.
	CreateEpair bool

	// JailInterfaceName is the desired interface name inside the jail (e.g. "eth0").
	// If empty, the epair name is kept.
	JailInterfaceName string

	// HostAddress is the IP address for the host-side bridge (e.g. "10.0.0.1/24").
	// Optional — set if you need the host to be on the same subnet. Requires Bridge
	// or CreateBridge.
	HostAddress string

	// JailAddress is the IP address to assign inside the jail (e.g. "10.0.0.2/24").
	JailAddress string

	// Gateway is the default gateway to set inside the jail (e.g. "10.0.0.1").
	Gateway string
}

// Setup configures VNET networking for a jail.
//
// cfg is updated in place with the names of any interfaces Setup created
// (EpairHost, EpairJail, Bridge), so the same value can be passed to Teardown.
// If Setup fails, the interfaces it created are destroyed again.
//
// Sequence:
//  1. Create epair (if CreateEpair)
//  2. Create bridge (if CreateBridge) and assign HostAddress to it
//  3. Add host-side epair to bridge
//  4. Bring up host-side epair
//  5. Move jail-side epair into the jail's vnet
//  6. Configure IP and route inside jail via exec.Shell
func Setup(cfg *NetworkConfig) (err error) {
	if cfg == nil || cfg.JailName == "" {
		name := ""
		if cfg != nil {
			name = cfg.JailName
		}
		return &jail.OperationError{Op: "vnet.setup", Jail: name, Err: jail.ErrInvalidArgument}
	}
	fail := func(step string, e error) error {
		return &jail.OperationError{Op: "vnet.setup", Jail: cfg.JailName, Err: fmt.Errorf("%s: %w", step, e)}
	}

	// Roll back whatever this call created if a later step fails.
	var createdEpair, createdBridge string
	defer func() {
		if err == nil {
			return
		}
		if createdEpair != "" {
			_ = epair.Destroy(createdEpair) // destroys both sides, even inside the jail
		}
		if createdBridge != "" {
			_ = bridge.Destroy(createdBridge)
		}
	}()

	// Step 1: Create epair if requested
	if cfg.CreateEpair {
		pair, err := epair.Create()
		if err != nil {
			return fail("create epair", err)
		}
		createdEpair = pair.A
		cfg.EpairHost = pair.A
		cfg.EpairJail = pair.B
	}

	if cfg.EpairHost == "" || cfg.EpairJail == "" {
		return fail("config", fmt.Errorf("%w: epair names must be set", jail.ErrInvalidArgument))
	}
	if cfg.HostAddress != "" && cfg.Bridge == "" && !cfg.CreateBridge {
		return fail("config", fmt.Errorf("%w: HostAddress requires Bridge or CreateBridge", jail.ErrInvalidArgument))
	}

	// Step 2: Create bridge if requested
	if cfg.CreateBridge {
		name, err := bridge.Create()
		if err != nil {
			return fail("create bridge", err)
		}
		createdBridge = name
		cfg.Bridge = name
		if err := bridge.Up(cfg.Bridge, true); err != nil {
			return fail("bridge up", err)
		}
	}

	if cfg.HostAddress != "" {
		if err := addAddress(cfg.Bridge, cfg.HostAddress); err != nil {
			return fail("assign host address", err)
		}
	}

	// Step 3: Add host-side epair to bridge (idempotent)
	if cfg.Bridge != "" {
		if err := bridge.AddMember(cfg.Bridge, cfg.EpairHost); err != nil {
			return fail("bridge add member", err)
		}
	}

	// Step 4: Bring up host-side epair
	if err := ifc.SetUp(cfg.EpairHost, true); err != nil {
		return fail("epair up", err)
	}

	// Step 5: Move jail-side epair into jail
	if err := jail.Update(cfg.JailName, jail.Config{
		VNETInterface: []string{cfg.EpairJail},
	}); err != nil {
		return fail("move interface to jail", err)
	}

	// Step 6: Configure networking inside the jail via exec
	ifName := cfg.EpairJail
	if cfg.JailInterfaceName != "" && cfg.JailInterfaceName != cfg.EpairJail {
		if err := jailexec.Shell(cfg.JailName, fmt.Sprintf("ifconfig %s name %s", cfg.EpairJail, cfg.JailInterfaceName)); err != nil {
			return fail("rename interface", err)
		}
		ifName = cfg.JailInterfaceName
	}

	if err := jailexec.Shell(cfg.JailName, fmt.Sprintf("ifconfig %s up", ifName)); err != nil {
		return fail("interface up", err)
	}

	if cfg.JailAddress != "" {
		family := "inet"
		if strings.Contains(cfg.JailAddress, ":") {
			family = "inet6"
		}
		if err := jailexec.Shell(cfg.JailName, fmt.Sprintf("ifconfig %s %s %s", ifName, family, cfg.JailAddress)); err != nil {
			return fail("assign "+family+" address", err)
		}
	}

	if cfg.Gateway != "" {
		cmd := fmt.Sprintf("route add default %s", cfg.Gateway)
		if strings.Contains(cfg.Gateway, ":") {
			cmd = fmt.Sprintf("route add -6 default %s", cfg.Gateway)
		}
		if err := jailexec.Shell(cfg.JailName, cmd); err != nil {
			return fail("add default route", err)
		}
	}

	return nil
}

// addAddress assigns a CIDR address to a host interface. An address that is
// already present is not an error.
func addAddress(iface, cidr string) error {
	addr, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return fmt.Errorf("%w: %v", jail.ErrInvalidArgument, err)
	}
	if v4 := addr.To4(); v4 != nil {
		err = ip.Add4(iface, v4, ipnet.Mask)
	} else {
		ones, _ := ipnet.Mask.Size()
		err = ip.Add6(iface, addr, ones)
	}
	if errors.Is(err, ifc.ErrExists) {
		return nil
	}
	return err
}

// Teardown removes VNET networking for a jail.
//
// Pass the same NetworkConfig that Setup filled in. The bridge is destroyed
// only if CreateBridge is set. Idempotent: missing interfaces/bridges are
// silently ignored.
func Teardown(cfg *NetworkConfig) error {
	if cfg == nil || cfg.JailName == "" {
		return &jail.OperationError{Op: "vnet.teardown", Err: jail.ErrInvalidArgument}
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
