//go:build freebsd

package jailops

import (
	"errors"
	"fmt"

	"github.com/zombocoder/go-freebsd-jail/internal/c"
	"github.com/zombocoder/go-freebsd-jail/internal/jailerr"
	"github.com/zombocoder/go-freebsd-jail/internal/param"
	"github.com/zombocoder/go-freebsd-jail/internal/types"
)

// Create creates a new jail. Returns the JID on success.
func Create(cfg types.Config) (int, error) {
	if err := validateCreateConfig(cfg); err != nil {
		return -1, &jailerr.OperationError{Op: "create", Jail: cfg.Name, Err: err}
	}

	params, err := param.ConfigToParams(cfg)
	if err != nil {
		return -1, &jailerr.OperationError{Op: "create", Jail: cfg.Name, Err: err}
	}

	jid, err := createJail(cfg, params)
	if err != nil {
		return -1, &jailerr.OperationError{Op: "create", Jail: cfg.Name, Err: err}
	}
	return jid, nil
}

// createJail mounts devfs (if requested), creates the jail and moves its
// VNET interfaces. On failure everything done so far is rolled back, so the
// caller never sees a half-configured jail.
func createJail(cfg types.Config, params []param.Param) (jid int, err error) {
	var devfsMountpoint string
	if cfg.MountDevFS {
		// Like jail(8), mount before creating so the jail starts with /dev.
		if devfsMountpoint, err = mountDevFS(cfg.Path, devfsRuleset(cfg)); err != nil {
			return -1, err
		}
	}
	defer func() {
		// Unmount exactly the (already validated) path that was mounted.
		if err != nil && devfsMountpoint != "" {
			_ = c.Unmount(devfsMountpoint, true)
		}
	}()

	jid, err = c.JailSet(paramsToCParams(params), c.FlagCreate)
	if err != nil {
		return -1, err
	}

	if err = moveInterfaces(jid, cfg.VNETInterface, false); err != nil {
		_ = c.JailRemove(jid)
		return -1, err
	}
	return jid, nil
}

// CreateOrUpdate creates a jail if it doesn't exist, or updates it if it does.
// On update, immutable params like 'path' are excluded.
func CreateOrUpdate(cfg types.Config) (int, error) {
	if err := validateCreateConfig(cfg); err != nil {
		return -1, &jailerr.OperationError{Op: "create_or_update", Jail: cfg.Name, Err: err}
	}

	params, err := param.ConfigToParams(cfg)
	if err != nil {
		return -1, &jailerr.OperationError{Op: "create_or_update", Jail: cfg.Name, Err: err}
	}

	// Try create first
	jid, err := createJail(cfg, params)
	if err == nil {
		return jid, nil
	}

	// If jail already exists, retry as update without immutable params
	if errors.Is(err, jailerr.ErrExists) {
		if cfg.MountDevFS {
			if _, err := mountDevFS(cfg.Path, devfsRuleset(cfg)); err != nil {
				return -1, &jailerr.OperationError{Op: "create_or_update", Jail: cfg.Name, Err: err}
			}
		}
		updateParams := filterImmutableParams(params)
		cparams := paramsToCParams(updateParams)
		jid, err = c.JailSet(cparams, c.FlagUpdate)
		if err != nil {
			return -1, &jailerr.OperationError{Op: "create_or_update", Jail: cfg.Name, Err: err}
		}
		// Interfaces moved by a previous call are no longer on the host;
		// skip them so repeated calls stay idempotent.
		if err := moveInterfaces(jid, cfg.VNETInterface, true); err != nil {
			return -1, &jailerr.OperationError{Op: "create_or_update", Jail: cfg.Name, Err: err}
		}
		return jid, nil
	}

	return -1, &jailerr.OperationError{Op: "create_or_update", Jail: cfg.Name, Err: err}
}

// filterImmutableParams removes params that cannot be changed after jail creation.
func filterImmutableParams(params []param.Param) []param.Param {
	// path and vnet can only be set when the jail is created; jail_set(2)
	// rejects them on update even when the value is unchanged.
	immutable := map[string]bool{"path": true, "vnet": true}
	filtered := make([]param.Param, 0, len(params))
	for _, p := range params {
		if !immutable[p.Name] {
			filtered = append(filtered, p)
		}
	}
	return filtered
}

func validateCreateConfig(cfg types.Config) error {
	if cfg.Name == "" {
		return &jailerr.ValidationError{Field: "name", Value: "", Err: jailerr.ErrInvalidArgument}
	}
	if cfg.Path == "" {
		return &jailerr.ValidationError{Field: "path", Value: "", Err: jailerr.ErrInvalidArgument}
	}
	if err := validateDevFS(cfg); err != nil {
		return err
	}
	if len(cfg.VNETInterface) > 0 && !cfg.VNET {
		return &jailerr.ValidationError{
			Field: "VNETInterface",
			Value: cfg.VNETInterface[0],
			Err:   fmt.Errorf("%w: VNETInterface requires VNET", jailerr.ErrInvalidArgument),
		}
	}
	return nil
}
