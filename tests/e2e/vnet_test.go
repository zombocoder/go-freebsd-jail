//go:build e2e && freebsd

package e2e

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/zombocoder/go-freebsd-ifc/epair"
	ifc "github.com/zombocoder/go-freebsd-ifc/if"

	jailexec "github.com/zombocoder/go-freebsd-jail/exec"
	"github.com/zombocoder/go-freebsd-jail/jail"
	"github.com/zombocoder/go-freebsd-jail/vnet"
)

// VNET tests use the host root as the jail path: networking inside the jail
// is configured with /bin/sh, /sbin/ifconfig and /sbin/route.
const vnetJailPath = "/"

func hostHasInterface(t *testing.T, name string) bool {
	t.Helper()
	_, err := ifc.Get(name)
	if err != nil && !errors.Is(err, ifc.ErrNotFound) {
		t.Fatalf("ifc.Get(%s): %v", name, err)
	}
	return err == nil
}

func TestVNETInterfaceOnCreate(t *testing.T) {
	skipIfNotRoot(t)

	pair, err := epair.Create()
	if err != nil {
		t.Fatalf("epair.Create: %v", err)
	}
	t.Cleanup(func() { _ = epair.Destroy(pair.A) })

	name := fmt.Sprintf("gotest-vnetif-%d", os.Getpid())
	t.Cleanup(func() { _ = jail.Remove(name) })

	cfg := jail.Config{
		Name:          name,
		Path:          vnetJailPath,
		Persist:       true,
		VNET:          true,
		VNETInterface: []string{pair.B},
	}
	if _, err := jail.Create(cfg); err != nil {
		t.Fatalf("Create with VNETInterface: %v", err)
	}

	if hostHasInterface(t, pair.B) {
		t.Errorf("%s still on host after Create", pair.B)
	}
	out, err := jailexec.Output(name, "/sbin/ifconfig", pair.B)
	if err != nil {
		t.Fatalf("%s not visible inside jail: %v: %s", pair.B, err, out)
	}

	// CreateOrUpdate with the same config must be idempotent even though
	// the interface has already left the host.
	if _, err := jail.CreateOrUpdate(cfg); err != nil {
		t.Fatalf("CreateOrUpdate (second call): %v", err)
	}
}

func TestVNETInterfaceRequiresVNET(t *testing.T) {
	skipIfNotRoot(t)

	name := fmt.Sprintf("gotest-novnet-%d", os.Getpid())
	t.Cleanup(func() { _ = jail.Remove(name) })

	_, err := jail.Create(jail.Config{
		Name:          name,
		Path:          vnetJailPath,
		Persist:       true,
		VNETInterface: []string{"epair999b"},
	})
	if !errors.Is(err, jail.ErrInvalidArgument) {
		t.Fatalf("expected ErrInvalidArgument, got %v", err)
	}
	if exists, _ := jail.Exists(name); exists {
		t.Fatal("jail must not be created when validation fails")
	}
}

func TestVNETCreateRollsBackOnMoveFailure(t *testing.T) {
	skipIfNotRoot(t)

	name := fmt.Sprintf("gotest-rollback-%d", os.Getpid())
	t.Cleanup(func() { _ = jail.Remove(name) })

	_, err := jail.Create(jail.Config{
		Name:          name,
		Path:          vnetJailPath,
		Persist:       true,
		VNET:          true,
		VNETInterface: []string{"nosuchif0"},
	})
	if err == nil {
		t.Fatal("expected error moving a nonexistent interface")
	}
	if exists, _ := jail.Exists(name); exists {
		t.Fatal("jail should have been rolled back after the interface move failed")
	}
}

func TestUnknownParamIsNotNotFound(t *testing.T) {
	skipIfNotRoot(t)

	name := fmt.Sprintf("gotest-badparam-%d", os.Getpid())
	t.Cleanup(func() { _ = jail.Remove(name) })

	_, err := jail.Create(jail.Config{
		Name:      name,
		Path:      vnetJailPath,
		Persist:   true,
		RawParams: map[string]string{"no.such.param": "1"},
	})
	if errors.Is(err, jail.ErrNotFound) {
		t.Fatalf("unknown parameter reported as ErrNotFound: %v", err)
	}
	if !errors.Is(err, jail.ErrInvalidArgument) {
		t.Fatalf("expected ErrInvalidArgument, got %v", err)
	}
}

func TestVNETSetupAndTeardown(t *testing.T) {
	skipIfNotRoot(t)

	name := fmt.Sprintf("gotest-vnetsetup-%d", os.Getpid())
	t.Cleanup(func() { _ = jail.Remove(name) })

	if _, err := jail.Create(jail.Config{
		Name:    name,
		Path:    vnetJailPath,
		Persist: true,
		VNET:    true,
	}); err != nil {
		t.Fatalf("Create: %v", err)
	}

	netCfg := vnet.NetworkConfig{
		JailName:          name,
		CreateBridge:      true,
		CreateEpair:       true,
		HostAddress:       "10.254.254.1/24",
		JailInterfaceName: "eth0",
		JailAddress:       "10.254.254.2/24",
		Gateway:           "10.254.254.1",
	}
	if err := vnet.Setup(&netCfg); err != nil {
		t.Fatalf("Setup: %v", err)
	}
	t.Cleanup(func() { _ = vnet.Teardown(&netCfg) })

	if netCfg.EpairHost == "" || netCfg.EpairJail == "" || netCfg.Bridge == "" {
		t.Fatalf("Setup did not report created interfaces: %+v", netCfg)
	}

	out, err := jailexec.Output(name, "/sbin/ifconfig", "eth0")
	if err != nil {
		t.Fatalf("ifconfig eth0 inside jail: %v", err)
	}
	if !strings.Contains(string(out), "10.254.254.2") {
		t.Errorf("jail address not configured:\n%s", out)
	}

	if out, err := jailexec.Output(name, "/sbin/ping", "-c", "1", "-t", "2", "10.254.254.1"); err != nil {
		t.Errorf("jail cannot reach host over the bridge: %v\n%s", err, out)
	}

	if err := vnet.Teardown(&netCfg); err != nil {
		t.Fatalf("Teardown: %v", err)
	}
	for _, ifname := range []string{netCfg.EpairHost, netCfg.Bridge} {
		if hostHasInterface(t, ifname) {
			t.Errorf("%s still exists after Teardown", ifname)
		}
	}
}

func TestVNETSetupRollsBackOnFailure(t *testing.T) {
	skipIfNotRoot(t)

	before, err := ifc.List()
	if err != nil {
		t.Fatalf("ifc.List: %v", err)
	}

	// The jail does not exist, so moving the interface fails after the
	// epair and bridge have been created.
	netCfg := vnet.NetworkConfig{
		JailName:     fmt.Sprintf("gotest-missing-%d", os.Getpid()),
		CreateBridge: true,
		CreateEpair:  true,
	}
	if err := vnet.Setup(&netCfg); err == nil {
		_ = vnet.Teardown(&netCfg)
		t.Fatal("expected Setup to fail for a nonexistent jail")
	}

	after, err := ifc.List()
	if err != nil {
		t.Fatalf("ifc.List: %v", err)
	}
	if len(after) != len(before) {
		t.Errorf("Setup leaked interfaces: %d before, %d after", len(before), len(after))
	}
}
