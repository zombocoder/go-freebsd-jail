// Package jail provides FreeBSD jail lifecycle management with native cgo/libjail bindings.
//
// This package allows Go applications to create, update, query, and remove FreeBSD jails
// using the kernel's jail_set(2), jail_get(2), and jail_remove(2) system calls via
// the libjail jailparam_* wrapper APIs.
//
// Most operations require root privileges. On non-FreeBSD systems, all functions
// return ErrNotSupported.
//
// WARNING: jail.Remove kills all processes in the target jail and removes child jails.
// This is a destructive, irreversible operation.
package jail
