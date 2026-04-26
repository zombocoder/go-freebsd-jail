//go:build freebsd

package jailops

import (
	"errors"
	"strconv"

	"github.com/zombocoder/go-freebsd-jail/internal/c"
	"github.com/zombocoder/go-freebsd-jail/internal/jailerr"
	"github.com/zombocoder/go-freebsd-jail/internal/param"
	"github.com/zombocoder/go-freebsd-jail/internal/types"
)

// Get retrieves information about a jail by name or JID string.
func Get(nameOrJID string) (*types.Info, error) {
	if nameOrJID == "" {
		return nil, &jailerr.OperationError{Op: "get", Jail: nameOrJID, Err: jailerr.ErrInvalidArgument}
	}

	key := buildKey(nameOrJID)
	names := param.StandardGetParams()

	jid, values, err := c.JailGet(key, names, 0)
	if err != nil {
		return nil, &jailerr.OperationError{Op: "get", Jail: nameOrJID, Err: err}
	}

	info := param.ValuesToInfo(jid, values)
	return info, nil
}

// List returns all active jails by iterating with the lastjid parameter.
func List() ([]types.Info, error) {
	var jails []types.Info
	lastjid := 0

	names := param.StandardGetParams()

	for {
		key := c.Param{Name: "lastjid", Value: strconv.Itoa(lastjid)}
		jid, values, err := c.JailGet(key, names, 0)
		if err != nil {
			if errors.Is(err, jailerr.ErrNotFound) {
				break // no more jails
			}
			return nil, &jailerr.OperationError{Op: "list", Jail: "", Err: err}
		}

		info := param.ValuesToInfo(jid, values)
		jails = append(jails, *info)
		lastjid = jid
	}

	return jails, nil
}

// Exists checks whether a jail exists by name or JID string.
func Exists(nameOrJID string) (bool, error) {
	_, err := Get(nameOrJID)
	if err != nil {
		if errors.Is(err, jailerr.ErrNotFound) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func buildKey(nameOrJID string) c.Param {
	if _, err := strconv.Atoi(nameOrJID); err == nil {
		return c.Param{Name: "jid", Value: nameOrJID}
	}
	return c.Param{Name: "name", Value: nameOrJID}
}
