//go:build e2e && freebsd

package e2e

import (
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/zombocoder/go-freebsd-jail/jail"
)

func skipIfNotRoot(t *testing.T) {
	t.Helper()
	if os.Getuid() != 0 {
		t.Skip("requires root privileges")
	}
}

func testJailPath(t *testing.T) string {
	t.Helper()
	// Use /rescue as a minimal rootfs for testing
	path := "/rescue"
	if _, err := os.Stat(path); err != nil {
		t.Skipf("test rootfs %s not available: %v", path, err)
	}
	return path
}

func TestCreateAndRemove(t *testing.T) {
	skipIfNotRoot(t)
	path := testJailPath(t)

	name := fmt.Sprintf("gotest-create-%d", os.Getpid())
	t.Cleanup(func() {
		_ = jail.Remove(name)
	})

	jid, err := jail.Create(jail.Config{
		Name:     name,
		Path:     path,
		Hostname: name + ".test",
		Persist:  true,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if jid <= 0 {
		t.Fatalf("expected positive JID, got %d", jid)
	}

	// Verify it exists
	exists, err := jail.Exists(name)
	if err != nil {
		t.Fatalf("Exists: %v", err)
	}
	if !exists {
		t.Fatal("jail should exist after creation")
	}

	// Remove
	if err := jail.Remove(name); err != nil {
		t.Fatalf("Remove: %v", err)
	}

	// Verify removed
	exists, err = jail.Exists(name)
	if err != nil {
		t.Fatalf("Exists after remove: %v", err)
	}
	if exists {
		t.Fatal("jail should not exist after removal")
	}
}

func TestGetJail(t *testing.T) {
	skipIfNotRoot(t)
	path := testJailPath(t)

	name := fmt.Sprintf("gotest-get-%d", os.Getpid())
	t.Cleanup(func() {
		_ = jail.Remove(name)
	})

	_, err := jail.Create(jail.Config{
		Name:     name,
		Path:     path,
		Hostname: name + ".test",
		Persist:  true,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	info, err := jail.Get(name)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}

	if info.Name != name {
		t.Errorf("Name: expected %q, got %q", name, info.Name)
	}
	if info.Path != path {
		t.Errorf("Path: expected %q, got %q", path, info.Path)
	}
	if info.Hostname != name+".test" {
		t.Errorf("Hostname: expected %q, got %q", name+".test", info.Hostname)
	}
	if info.State != jail.StateActive {
		t.Errorf("State: expected %q, got %q", jail.StateActive, info.State)
	}
}

func TestListContainsJail(t *testing.T) {
	skipIfNotRoot(t)
	path := testJailPath(t)

	name := fmt.Sprintf("gotest-list-%d", os.Getpid())
	t.Cleanup(func() {
		_ = jail.Remove(name)
	})

	_, err := jail.Create(jail.Config{
		Name:     name,
		Path:     path,
		Hostname: name + ".test",
		Persist:  true,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	jails, err := jail.List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}

	found := false
	for _, j := range jails {
		if j.Name == name {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("jail %q not found in List() results", name)
	}
}

func TestCreateOrUpdate(t *testing.T) {
	skipIfNotRoot(t)
	path := testJailPath(t)

	name := fmt.Sprintf("gotest-cou-%d", os.Getpid())
	t.Cleanup(func() {
		_ = jail.Remove(name)
	})

	// Create
	jid1, err := jail.CreateOrUpdate(jail.Config{
		Name:     name,
		Path:     path,
		Hostname: name + ".test",
		Persist:  true,
	})
	if err != nil {
		t.Fatalf("CreateOrUpdate (create): %v", err)
	}

	// Update
	jid2, err := jail.CreateOrUpdate(jail.Config{
		Name:     name,
		Path:     path,
		Hostname: name + ".updated",
		Persist:  true,
	})
	if err != nil {
		t.Fatalf("CreateOrUpdate (update): %v", err)
	}

	if jid1 != jid2 {
		t.Errorf("JID changed: %d -> %d", jid1, jid2)
	}

	info, err := jail.Get(name)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if info.Hostname != name+".updated" {
		t.Errorf("Hostname not updated: expected %q, got %q", name+".updated", info.Hostname)
	}
}

func TestCreateDuplicate(t *testing.T) {
	skipIfNotRoot(t)
	path := testJailPath(t)

	name := fmt.Sprintf("gotest-dup-%d", os.Getpid())
	t.Cleanup(func() {
		_ = jail.Remove(name)
	})

	_, err := jail.Create(jail.Config{
		Name:     name,
		Path:     path,
		Hostname: name + ".test",
		Persist:  true,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	_, err = jail.Create(jail.Config{
		Name:     name,
		Path:     path,
		Hostname: name + ".test",
		Persist:  true,
	})
	if err == nil {
		t.Fatal("expected error creating duplicate jail")
	}
	if !errors.Is(err, jail.ErrExists) {
		t.Errorf("expected ErrExists, got: %v", err)
	}
}

func TestRemoveIdempotent(t *testing.T) {
	skipIfNotRoot(t)

	// Removing a non-existent jail should not error
	err := jail.Remove("gotest-nonexistent-jail-12345")
	if err != nil {
		t.Errorf("Remove non-existent: expected nil, got %v", err)
	}
}
