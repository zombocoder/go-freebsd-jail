package exec

import (
	"io"

	"github.com/zombocoder/go-freebsd-jail/internal/execops"
)

// Command describes a command to execute inside a jail.
type Command struct {
	// Jail is the jail name or JID string.
	Jail string

	// Path is the executable path inside the jail.
	Path string

	// Args are the command arguments (not including the command itself).
	Args []string

	// Env specifies the environment of the process.
	Env []string

	// Dir specifies the working directory.
	Dir string

	// Stdout, Stderr, and Stdin configure the process I/O.
	Stdout io.Writer
	Stderr io.Writer
	Stdin  io.Reader
}

// Run executes a command inside a jail and waits for it to complete.
// Uses /usr/sbin/jexec to safely execute without attaching the Go runtime.
func Run(cmd Command) error {
	opts := &execops.Options{
		Stdin:  cmd.Stdin,
		Stdout: cmd.Stdout,
		Stderr: cmd.Stderr,
		Env:    cmd.Env,
		Dir:    cmd.Dir,
	}
	return execops.Run(cmd.Jail, cmd.Path, cmd.Args, opts)
}

// Output executes a command inside a jail and returns its standard output.
func Output(jail string, path string, args ...string) ([]byte, error) {
	return execops.Output(jail, path, args)
}

// Shell executes a shell command inside a jail via /bin/sh -c.
func Shell(jail string, command string) error {
	return execops.Shell(jail, command)
}
