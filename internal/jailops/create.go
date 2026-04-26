//go:build freebsd

package jailops

import (
	"errors"

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

	cparams := paramsToCParams(params)
	jid, err := c.JailSet(cparams, c.FlagCreate)
	if err != nil {
		return -1, &jailerr.OperationError{Op: "create", Jail: cfg.Name, Err: err}
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
	cparams := paramsToCParams(params)
	jid, err := c.JailSet(cparams, c.FlagCreate)
	if err == nil {
		return jid, nil
	}

	// If jail already exists, retry as update without immutable params
	if errors.Is(err, jailerr.ErrExists) {
		updateParams := filterImmutableParams(params)
		cparams = paramsToCParams(updateParams)
		jid, err = c.JailSet(cparams, c.FlagUpdate)
		if err != nil {
			return -1, &jailerr.OperationError{Op: "create_or_update", Jail: cfg.Name, Err: err}
		}
		return jid, nil
	}

	return -1, &jailerr.OperationError{Op: "create_or_update", Jail: cfg.Name, Err: err}
}

// filterImmutableParams removes params that cannot be changed after jail creation.
func filterImmutableParams(params []param.Param) []param.Param {
	immutable := map[string]bool{"path": true}
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
	return nil
}
