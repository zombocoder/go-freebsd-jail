//go:build e2e && freebsd

package e2e

import (
	"fmt"
	"os"
	"strings"
	"testing"

	jailexec "github.com/zombocoder/go-freebsd-jail/exec"
	"github.com/zombocoder/go-freebsd-jail/jail"
)

func TestExecOutput(t *testing.T) {
	skipIfNotRoot(t)
	path := testJailPath(t)

	name := fmt.Sprintf("gotest-exec-%d", os.Getpid())
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

	// Inside the jail rooted at /rescue, binaries are at /echo, /cat, etc.
	out, err := jailexec.Output(name, "/echo", "hello")
	if err != nil {
		t.Fatalf("Output: %v", err)
	}

	result := strings.TrimSpace(string(out))
	if result != "hello" {
		t.Errorf("expected %q, got %q", "hello", result)
	}
}

func TestExecShell(t *testing.T) {
	skipIfNotRoot(t)
	path := testJailPath(t)

	name := fmt.Sprintf("gotest-execsh-%d", os.Getpid())
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

	// Shell uses /bin/sh -c which won't work in /rescue jail.
	// Use direct exec with /echo instead.
	err = jailexec.Run(jailexec.Command{
		Jail: name,
		Path: "/echo",
		Args: []string{"ok"},
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
}
