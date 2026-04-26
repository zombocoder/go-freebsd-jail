// Package exec provides command execution inside FreeBSD jails.
//
// Commands are executed using /usr/sbin/jexec, which safely forks a child process,
// attaches it to the target jail, and executes the command. This avoids the dangerous
// pattern of calling jail_attach(2) in the Go runtime process, which would change
// the jail context for all goroutines.
//
// On non-FreeBSD systems, all functions return jail.ErrNotSupported.
package exec
