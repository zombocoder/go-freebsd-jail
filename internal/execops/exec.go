//go:build freebsd

package execops

import (
	"bytes"
	"fmt"
	"io"
	"os/exec"

	"github.com/zombocoder/go-freebsd-jail/internal/jailerr"
)

const jexecPath = "/usr/sbin/jexec"

// Options configures command execution inside a jail.
type Options struct {
	Stdin  io.Reader
	Stdout io.Writer
	Stderr io.Writer
	Env    []string
	Dir    string
}

// Run executes a command inside a jail and waits for it to complete.
func Run(jailName string, path string, args []string, opts *Options) error {
	if jailName == "" {
		return &jailerr.OperationError{Op: "exec", Jail: jailName, Err: jailerr.ErrInvalidArgument}
	}
	if path == "" {
		return &jailerr.OperationError{Op: "exec", Jail: jailName, Err: fmt.Errorf("%w: command path is empty", jailerr.ErrInvalidArgument)}
	}

	cmd := buildCmd(jailName, path, args, opts)
	if err := cmd.Run(); err != nil {
		return &jailerr.OperationError{Op: "exec", Jail: jailName, Err: err}
	}
	return nil
}

// Output executes a command inside a jail and returns its stdout.
func Output(jailName string, path string, args []string) ([]byte, error) {
	if jailName == "" {
		return nil, &jailerr.OperationError{Op: "exec", Jail: jailName, Err: jailerr.ErrInvalidArgument}
	}
	if path == "" {
		return nil, &jailerr.OperationError{Op: "exec", Jail: jailName, Err: fmt.Errorf("%w: command path is empty", jailerr.ErrInvalidArgument)}
	}

	cmd := buildCmd(jailName, path, args, nil)
	out, err := cmd.Output()
	if err != nil {
		return out, &jailerr.OperationError{Op: "exec", Jail: jailName, Err: err}
	}
	return out, nil
}

// Shell executes a shell command inside a jail via /bin/sh -c.
func Shell(jailName string, command string) error {
	return Run(jailName, "/bin/sh", []string{"-c", command}, nil)
}

// ShellOutput executes a shell command and returns its stdout.
func ShellOutput(jailName string, command string) ([]byte, error) {
	var buf bytes.Buffer
	opts := &Options{Stdout: &buf}
	cmd := buildCmd(jailName, "/bin/sh", []string{"-c", command}, opts)
	if err := cmd.Run(); err != nil {
		return buf.Bytes(), &jailerr.OperationError{Op: "exec", Jail: jailName, Err: err}
	}
	return buf.Bytes(), nil
}

func buildCmd(jailName string, path string, args []string, opts *Options) *exec.Cmd {
	cmdArgs := make([]string, 0, 2+len(args))
	cmdArgs = append(cmdArgs, jailName, path)
	cmdArgs = append(cmdArgs, args...)

	cmd := exec.Command(jexecPath, cmdArgs...)

	if opts != nil {
		cmd.Stdin = opts.Stdin
		cmd.Stdout = opts.Stdout
		cmd.Stderr = opts.Stderr
		cmd.Env = opts.Env
		cmd.Dir = opts.Dir
	}

	return cmd
}
