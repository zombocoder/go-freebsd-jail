//go:build !freebsd

package jailops

import (
	"github.com/zombocoder/go-freebsd-jail/internal/jailerr"
	"github.com/zombocoder/go-freebsd-jail/internal/types"
)

func Create(cfg types.Config) (int, error) {
	return -1, jailerr.ErrNotSupported
}

func CreateOrUpdate(cfg types.Config) (int, error) {
	return -1, jailerr.ErrNotSupported
}

func Update(nameOrJID string, cfg types.Config) error {
	return jailerr.ErrNotSupported
}

func Get(nameOrJID string) (*types.Info, error) {
	return nil, jailerr.ErrNotSupported
}

func List() ([]types.Info, error) {
	return nil, jailerr.ErrNotSupported
}

func Exists(nameOrJID string) (bool, error) {
	return false, jailerr.ErrNotSupported
}

func Remove(nameOrJID string) error {
	return jailerr.ErrNotSupported
}
