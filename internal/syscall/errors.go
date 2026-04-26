//go:build freebsd

package syscall

import (
	"fmt"
	"syscall"

	"github.com/zombocoder/go-freebsd-jail/internal/jailerr"
)

// MapErrno maps a FreeBSD errno to a typed jail error.
func MapErrno(errno syscall.Errno, errmsg string) error {
	base := mapErrnoBase(errno)
	if errmsg != "" {
		return fmt.Errorf("%w: %s", base, errmsg)
	}
	return base
}

func mapErrnoBase(errno syscall.Errno) error {
	switch errno {
	case syscall.EPERM, syscall.EACCES:
		return jailerr.ErrPermission
	case syscall.ENOENT:
		return jailerr.ErrNotFound
	case syscall.EEXIST:
		return jailerr.ErrExists
	case syscall.EINVAL:
		return jailerr.ErrInvalidArgument
	case syscall.EBUSY:
		return jailerr.ErrBusy
	case syscall.ENOMEM:
		return jailerr.ErrNoMemory
	default:
		return fmt.Errorf("syscall error %d: %s", errno, errno.Error())
	}
}
