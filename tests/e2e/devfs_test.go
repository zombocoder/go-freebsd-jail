//go:build e2e && freebsd

package e2e

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"syscall"
	"testing"

	"github.com/zombocoder/go-freebsd-jail/jail"
)

// devfsJailRoot returns an empty jail root containing only a dev/ directory.
func devfsJailRoot(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, "dev"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		// Safety net: never leave a devfs mounted under the temp dir,
		// otherwise TempDir cleanup fails.
		if isDevFS(t, filepath.Join(root, "dev")) {
			_ = syscall.Unmount(filepath.Join(root, "dev"), 0x80000) // MNT_FORCE
		}
	})
	return root
}

func isDevFS(t *testing.T, path string) bool {
	t.Helper()
	var st syscall.Statfs_t
	if err := syscall.Statfs(path, &st); err != nil {
		t.Fatalf("statfs %s: %v", path, err)
	}
	var fstype, mnton []byte
	for _, c := range st.Fstypename {
		if c == 0 {
			break
		}
		fstype = append(fstype, byte(c))
	}
	for _, c := range st.Mntonname {
		if c == 0 {
			break
		}
		mnton = append(mnton, byte(c))
	}
	return string(fstype) == "devfs" && string(mnton) == path
}

func TestMountDevFS(t *testing.T) {
	skipIfNotRoot(t)
	root := devfsJailRoot(t)
	dev := filepath.Join(root, "dev")

	name := fmt.Sprintf("gotest-devfs-%d", os.Getpid())
	t.Cleanup(func() { _ = jail.Remove(name) })

	cfg := jail.Config{Name: name, Path: root, Persist: true, MountDevFS: true}
	if _, err := jail.Create(cfg); err != nil {
		t.Fatalf("Create with MountDevFS: %v", err)
	}

	if !isDevFS(t, dev) {
		t.Fatalf("devfs not mounted on %s", dev)
	}
	if _, err := os.Stat(filepath.Join(dev, "null")); err != nil {
		t.Errorf("/dev/null missing in jail devfs: %v", err)
	}
	// Ruleset 4 (devfsrules_jail) hides raw memory devices.
	if _, err := os.Stat(filepath.Join(dev, "mem")); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("/dev/mem visible in jail devfs; ruleset 4 not applied (err=%v)", err)
	}

	// Idempotent: no second devfs stacked on top.
	if _, err := jail.CreateOrUpdate(cfg); err != nil {
		t.Fatalf("CreateOrUpdate: %v", err)
	}
	if err := syscall.Unmount(dev, 0); err != nil {
		t.Fatalf("unmount first devfs: %v", err)
	}
	if isDevFS(t, dev) {
		t.Fatal("CreateOrUpdate stacked a second devfs mount")
	}

	// Update with MountDevFS remounts it, looking up the path itself.
	if err := jail.Update(name, jail.Config{MountDevFS: true}); err != nil {
		t.Fatalf("Update with MountDevFS: %v", err)
	}
	if !isDevFS(t, dev) {
		t.Fatal("Update did not mount devfs")
	}

	if err := jail.Remove(name); err != nil {
		t.Fatalf("Remove: %v", err)
	}
	if isDevFS(t, dev) {
		t.Error("devfs still mounted after Remove")
	}
}

func TestMountDevFSCustomRuleset(t *testing.T) {
	skipIfNotRoot(t)
	root := devfsJailRoot(t)
	dev := filepath.Join(root, "dev")

	name := fmt.Sprintf("gotest-devfsrs-%d", os.Getpid())
	t.Cleanup(func() { _ = jail.Remove(name) })

	ruleset := 0 // no rules: everything visible
	if _, err := jail.Create(jail.Config{
		Name: name, Path: root, Persist: true, MountDevFS: true, DevFSRuleset: &ruleset,
	}); err != nil {
		t.Fatalf("Create: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dev, "mem")); err != nil {
		t.Errorf("ruleset 0 should expose /dev/mem: %v", err)
	}
}

func TestMountDevFSRejectsHostRoot(t *testing.T) {
	skipIfNotRoot(t)

	name := fmt.Sprintf("gotest-devfsroot-%d", os.Getpid())
	t.Cleanup(func() { _ = jail.Remove(name) })

	_, err := jail.Create(jail.Config{Name: name, Path: "/", Persist: true, MountDevFS: true})
	if !errors.Is(err, jail.ErrInvalidArgument) {
		t.Fatalf("expected ErrInvalidArgument for MountDevFS on /, got %v", err)
	}
}

func TestMountDevFSRollback(t *testing.T) {
	skipIfNotRoot(t)
	root := devfsJailRoot(t)

	name := fmt.Sprintf("gotest-devfsrb-%d", os.Getpid())
	t.Cleanup(func() { _ = jail.Remove(name) })

	// Moving a nonexistent interface fails after devfs is mounted.
	_, err := jail.Create(jail.Config{
		Name: name, Path: root, Persist: true, MountDevFS: true,
		VNET: true, VNETInterface: []string{"nosuchif0"},
	})
	if err == nil {
		t.Fatal("expected Create to fail")
	}
	if isDevFS(t, filepath.Join(root, "dev")) {
		t.Error("devfs left mounted after failed Create")
	}
}

func TestMountDevFSMissingDevDir(t *testing.T) {
	skipIfNotRoot(t)
	root := t.TempDir() // no dev/ inside

	name := fmt.Sprintf("gotest-devfsnodir-%d", os.Getpid())
	t.Cleanup(func() { _ = jail.Remove(name) })

	if _, err := jail.Create(jail.Config{Name: name, Path: root, Persist: true, MountDevFS: true}); err == nil {
		t.Fatal("expected error when <path>/dev does not exist")
	}
	if exists, _ := jail.Exists(name); exists {
		t.Error("jail created despite devfs mount failure")
	}
}

// A jail root is often writable from inside the jail. If <path>/dev is a
// symlink, mounting on it would put a devfs over the symlink's target on the
// host, so it must be refused (jail(8) refuses it too).
func TestMountDevFSRefusesSymlink(t *testing.T) {
	skipIfNotRoot(t)
	root := t.TempDir()
	target := t.TempDir()
	if err := os.Symlink(target, filepath.Join(root, "dev")); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if isDevFS(t, target) {
			_ = syscall.Unmount(target, 0x80000) // MNT_FORCE
		}
	})

	name := fmt.Sprintf("gotest-devfslink-%d", os.Getpid())
	t.Cleanup(func() { _ = jail.Remove(name) })

	_, err := jail.Create(jail.Config{Name: name, Path: root, Persist: true, MountDevFS: true})
	if !errors.Is(err, jail.ErrInvalidArgument) {
		t.Errorf("expected ErrInvalidArgument for symlinked dev, got %v", err)
	}
	if isDevFS(t, target) {
		t.Fatalf("devfs was mounted on symlink target %s", target)
	}
	if exists, _ := jail.Exists(name); exists {
		t.Error("jail created despite refused devfs mount")
	}
}

// A jail path that is a symlink to / must be treated like "/": no devfs mount
// on, and above all no unmount of, the host's /dev.
func TestMountDevFSRefusesPathResolvingToRoot(t *testing.T) {
	skipIfNotRoot(t)
	link := filepath.Join(t.TempDir(), "rootlink")
	if err := os.Symlink("/", link); err != nil {
		t.Fatal(err)
	}

	name := fmt.Sprintf("gotest-devfsrootlink-%d", os.Getpid())
	t.Cleanup(func() { _ = jail.Remove(name) })

	_, err := jail.Create(jail.Config{Name: name, Path: link, Persist: true, MountDevFS: true})
	if !errors.Is(err, jail.ErrInvalidArgument) {
		t.Errorf("expected ErrInvalidArgument, got %v", err)
	}

	// Remove of a jail rooted at / via a symlink must leave the host /dev alone.
	if _, err := jail.Create(jail.Config{Name: name, Path: link, Persist: true}); err != nil {
		t.Fatalf("Create without devfs: %v", err)
	}
	if err := jail.Remove(name); err != nil {
		t.Fatalf("Remove: %v", err)
	}
	if !isDevFS(t, "/dev") {
		t.Fatal("host /dev was unmounted")
	}
}
