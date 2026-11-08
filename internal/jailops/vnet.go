//go:build freebsd

package jailops

import (
	"errors"
	"fmt"

	ifc "github.com/zombocoder/go-freebsd-ifc/if"
)

// moveInterfaces moves host interfaces into the vnet of jail jid.
//
// vnet.interface is a jail(8) pseudo-parameter that the kernel does not
// accept, so this does what jail(8) does: SIOCSIFVNET per interface.
//
// With skipMissing, interfaces that are not on the host are ignored; this
// keeps CreateOrUpdate idempotent once an interface has already been moved.
func moveInterfaces(jid int, ifaces []string, skipMissing bool) error {
	for _, name := range ifaces {
		if skipMissing {
			if _, err := ifc.Get(name); errors.Is(err, ifc.ErrNotFound) {
				continue
			}
		}
		if err := ifc.MoveToVNET(name, jid); err != nil {
			return fmt.Errorf("move interface %s to jail %d: %w", name, jid, err)
		}
	}
	return nil
}
