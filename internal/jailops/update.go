//go:build freebsd

package jailops

import (
	"github.com/zombocoder/go-freebsd-jail/internal/c"
	"github.com/zombocoder/go-freebsd-jail/internal/jailerr"
	"github.com/zombocoder/go-freebsd-jail/internal/param"
	"github.com/zombocoder/go-freebsd-jail/internal/types"
)

// Update modifies parameters of an existing jail.
func Update(nameOrJID string, cfg types.Config) error {
	if nameOrJID == "" {
		return &jailerr.OperationError{Op: "update", Jail: nameOrJID, Err: jailerr.ErrInvalidArgument}
	}

	if cfg.Name == "" {
		cfg.Name = nameOrJID
	}

	params, err := param.ConfigToParams(cfg)
	if err != nil {
		return &jailerr.OperationError{Op: "update", Jail: nameOrJID, Err: err}
	}

	cparams := paramsToCParams(params)
	_, err = c.JailSet(cparams, c.FlagUpdate)
	if err != nil {
		return &jailerr.OperationError{Op: "update", Jail: nameOrJID, Err: err}
	}
	return nil
}
