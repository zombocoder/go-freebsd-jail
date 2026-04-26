//go:build !freebsd

package execops

import (
	"io"

	"github.com/zombocoder/go-freebsd-jail/internal/jailerr"
)

type Options struct {
	Stdin  io.Reader
	Stdout io.Writer
	Stderr io.Writer
	Env    []string
	Dir    string
}

func Run(jailName string, path string, args []string, opts *Options) error {
	return jailerr.ErrNotSupported
}

func Output(jailName string, path string, args []string) ([]byte, error) {
	return nil, jailerr.ErrNotSupported
}

func Shell(jailName string, command string) error {
	return jailerr.ErrNotSupported
}

func ShellOutput(jailName string, command string) ([]byte, error) {
	return nil, jailerr.ErrNotSupported
}
