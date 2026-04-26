//go:build freebsd

package syscall

import "syscall"

// Common errno values used in jail operations.
const (
	EPERM  = syscall.EPERM
	ENOENT = syscall.ENOENT
	EEXIST = syscall.EEXIST
	EINVAL = syscall.EINVAL
	EBUSY  = syscall.EBUSY
	ENOMEM = syscall.ENOMEM
	EACCES = syscall.EACCES
)
