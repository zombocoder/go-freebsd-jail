//go:build !freebsd

package c

import "github.com/zombocoder/go-freebsd-jail/internal/jailerr"

// Param represents a jail parameter for the C layer.
type Param struct {
	Name   string
	Value  string
	IsBool bool
}

// Jail operation flags.
const (
	FlagCreate = 0x01
	FlagUpdate = 0x02
)

func JailSet(params []Param, flags int) (int, error) {
	return -1, jailerr.ErrNotSupported
}

func JailGet(key Param, names []string, flags int) (int, map[string]string, error) {
	return -1, nil, jailerr.ErrNotSupported
}

func JailRemove(jid int) error {
	return jailerr.ErrNotSupported
}

func JailAttach(jid int) error {
	return jailerr.ErrNotSupported
}
