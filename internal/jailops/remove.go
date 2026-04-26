//go:build freebsd

package jailops

import (
	"errors"

	"github.com/zombocoder/go-freebsd-jail/internal/c"
	"github.com/zombocoder/go-freebsd-jail/internal/jailerr"
)

// Remove removes a jail by name or JID string.
// Idempotent: removing a non-existent jail returns nil.
// WARNING: This kills all processes in the jail and removes child jails.
func Remove(nameOrJID string) error {
	if nameOrJID == "" {
		return &jailerr.OperationError{Op: "remove", Jail: nameOrJID, Err: jailerr.ErrInvalidArgument}
	}

	info, err := Get(nameOrJID)
	if err != nil {
		if errors.Is(err, jailerr.ErrNotFound) {
			return nil // idempotent
		}
		return &jailerr.OperationError{Op: "remove", Jail: nameOrJID, Err: err}
	}

	err = c.JailRemove(int(info.JID))
	if err != nil {
		if errors.Is(err, jailerr.ErrNotFound) {
			return nil // raced with another removal
		}
		return &jailerr.OperationError{Op: "remove", Jail: nameOrJID, Err: err}
	}
	return nil
}
