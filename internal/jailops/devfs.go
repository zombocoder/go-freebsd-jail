//go:build freebsd

package jailops

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"syscall"

	"github.com/zombocoder/go-freebsd-jail/internal/c"
	"github.com/zombocoder/go-freebsd-jail/internal/jailerr"
	"github.com/zombocoder/go-freebsd-jail/internal/types"
)

// defaultDevFSRuleset is the ruleset jail(8) applies for mount.devfs when
// devfs_ruleset is not set: 4, "devfsrules_jail" from /etc/defaults/devfs.rules.
const defaultDevFSRuleset = 4

// devPath returns <jailPath>/dev with the jail root's symlinks resolved,
// and refuses a dev entry that is a symlink or not a directory.
//
// The jail root is typically writable from inside the jail. nmount(2) and
// unmount(2) follow symlinks, so a dev -> /etc symlink planted by the jail
// would otherwise put a devfs over (or unmount) a host directory. jail(8)
// refuses such mount points as well. A dev entry that does not exist is
// returned as is; mounting on it then fails with ENOENT.
//
// The check runs before the mount, so a process already inside the jail
// could still swap the directory in between (on CreateOrUpdate/Update of a
// running jail); on Create there is no such process yet.
func devPath(jailPath string) (string, error) {
	root, err := filepath.EvalSymlinks(jailPath)
	if err != nil {
		return "", fmt.Errorf("resolve jail path %s: %w", jailPath, err)
	}
	if root == "/" {
		return "", &jailerr.ValidationError{
			Field: "MountDevFS",
			Value: jailPath,
			Err:   fmt.Errorf("%w: jail path resolves to /; refusing to touch the host's /dev", jailerr.ErrInvalidArgument),
		}
	}
	dev := filepath.Join(root, "dev")

	fi, err := os.Lstat(dev)
	switch {
	case errors.Is(err, os.ErrNotExist):
		return dev, nil
	case err != nil:
		return "", err
	case fi.Mode()&os.ModeSymlink != 0:
		return "", &jailerr.ValidationError{
			Field: "MountDevFS",
			Value: dev,
			Err:   fmt.Errorf("%w: %s is a symlink; refusing to mount devfs through it", jailerr.ErrInvalidArgument, dev),
		}
	case !fi.IsDir():
		return "", &jailerr.ValidationError{
			Field: "MountDevFS",
			Value: dev,
			Err:   fmt.Errorf("%w: %s is not a directory", jailerr.ErrInvalidArgument, dev),
		}
	}
	return dev, nil
}

func devfsRuleset(cfg types.Config) int {
	if cfg.DevFSRuleset != nil {
		return *cfg.DevFSRuleset
	}
	return defaultDevFSRuleset
}

// isDevFSMountpoint reports whether a devfs is mounted exactly on path.
func isDevFSMountpoint(path string) (bool, error) {
	var st syscall.Statfs_t
	if err := syscall.Statfs(path, &st); err != nil {
		return false, err
	}
	return cstr(st.Fstypename[:]) == "devfs" && cstr(st.Mntonname[:]) == path, nil
}

func cstr(b []int8) string {
	buf := make([]byte, 0, len(b))
	for _, c := range b {
		if c == 0 {
			break
		}
		buf = append(buf, byte(c))
	}
	return string(buf)
}

// mountDevFS mounts devfs on <jailPath>/dev, like jail(8) mount.devfs.
// It returns the mount point it used, if it mounted anything; an existing
// devfs mount is reused and reported as "".
func mountDevFS(jailPath string, ruleset int) (string, error) {
	dev, err := devPath(jailPath)
	if err != nil {
		return "", err
	}
	if mounted, err := isDevFSMountpoint(dev); err != nil {
		return "", fmt.Errorf("mount devfs on %s: %w", dev, err)
	} else if mounted {
		return "", nil
	}
	if err := c.MountDevFS(dev, ruleset); err != nil {
		return "", fmt.Errorf("mount devfs on %s: %w", dev, err)
	}
	return dev, nil
}

// unmountDevFS unmounts the devfs on <jailPath>/dev if it is no longer used
// by any other active jail with the same path. The host's /dev is never touched.
func unmountDevFS(jailPath string) error {
	if filepath.Clean(jailPath) == "/" {
		return nil
	}
	dev, err := devPath(jailPath)
	if err != nil {
		return nil // symlinked or replaced dev: never unmount through it
	}
	mounted, err := isDevFSMountpoint(dev)
	if err != nil || !mounted {
		return nil
	}

	jails, err := List()
	if err != nil {
		return err
	}
	for _, j := range jails {
		if filepath.Clean(j.Path) == filepath.Clean(jailPath) {
			return nil // still used by another jail
		}
	}

	err = c.Unmount(dev, false)
	if errors.Is(err, jailerr.ErrBusy) {
		// Processes of a dying jail can still hold device nodes open.
		err = c.Unmount(dev, true)
	}
	if err != nil {
		return fmt.Errorf("unmount devfs on %s: %w", dev, err)
	}
	return nil
}

func validateDevFS(cfg types.Config) error {
	if !cfg.MountDevFS {
		return nil
	}
	if cfg.Path != "" && filepath.Clean(cfg.Path) == "/" {
		return &jailerr.ValidationError{
			Field: "MountDevFS",
			Value: cfg.Path,
			Err:   fmt.Errorf("%w: refusing to mount devfs over the host's /dev", jailerr.ErrInvalidArgument),
		}
	}
	if cfg.DevFSRuleset != nil && *cfg.DevFSRuleset < 0 {
		return &jailerr.ValidationError{
			Field: "DevFSRuleset",
			Value: fmt.Sprint(*cfg.DevFSRuleset),
			Err:   jailerr.ErrInvalidArgument,
		}
	}
	return nil
}
