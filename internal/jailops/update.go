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

	if cfg.MountDevFS {
		if err := ensureDevFS(nameOrJID, cfg); err != nil {
			return &jailerr.OperationError{Op: "update", Jail: nameOrJID, Err: err}
		}
	}

	cparams := paramsToCParams(params)
	jid, err := c.JailSet(cparams, c.FlagUpdate)
	if err != nil {
		return &jailerr.OperationError{Op: "update", Jail: nameOrJID, Err: err}
	}

	if err := moveInterfaces(jid, cfg.VNETInterface, false); err != nil {
		return &jailerr.OperationError{Op: "update", Jail: nameOrJID, Err: err}
	}
	return nil
}

// ensureDevFS mounts devfs on the existing jail's /dev if it is not mounted.
func ensureDevFS(nameOrJID string, cfg types.Config) error {
	if cfg.Path == "" {
		info, err := Get(nameOrJID)
		if err != nil {
			return err
		}
		cfg.Path = info.Path
	}
	if err := validateDevFS(cfg); err != nil {
		return err
	}
	_, err := mountDevFS(cfg.Path, devfsRuleset(cfg))
	return err
}
