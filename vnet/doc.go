// Package vnet provides higher-level helpers for VNET jail networking.
//
// It integrates with github.com/zombocoder/go-freebsd-ifc for epair and bridge
// management, and uses the exec package to configure networking inside the jail's
// virtual network stack.
//
// VNET jails have their own isolated network stack. Interfaces are moved into
// the jail via the vnet.interface kernel parameter. Once moved, the interface
// is no longer visible on the host.
//
// On non-FreeBSD systems, all functions return jail.ErrNotSupported.
package vnet
