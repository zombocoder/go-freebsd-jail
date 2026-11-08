# Changelog

All notable changes to this project are documented here. The format is based
on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and the project
follows [Semantic Versioning](https://semver.org/).

## [Unreleased]

### Changed

- **Breaking:** `vnet.Setup` and `vnet.Teardown` take a `*vnet.NetworkConfig`.
  `Setup` fills in the names of the interfaces it created (`EpairHost`,
  `EpairJail`, `Bridge`) so the same value can be passed to `Teardown`;
  previously the names were lost and `Teardown` could not remove them.
- Requires go-freebsd-ifc v1.1.0 or later (`MoveToVNET`).

### Added

- `vnet.Setup` assigns `HostAddress` to the bridge (the field was documented
  but ignored) and rolls back the epair and bridge it created when a later
  step fails.
- `MountDevFS` mounts devfs on `<Path>/dev` with `nmount(2)` before the jail is
  created, applying `DevFSRuleset` (default 4, `devfsrules_jail`), like
  jail(8)'s `mount.devfs`. `Remove` unmounts it once no other jail uses the
  path. Like jail(8), the mount point must be a real directory inside the
  jail root: a `dev` symlink (which the jail itself could plant to get a devfs
  mounted over a host directory) and paths resolving to `/` are refused.
- `Create` and `CreateOrUpdate` roll back (remove the jail, unmount devfs)
  when moving an interface or mounting devfs fails.
- E2E tests for VNET jails, `vnet.Setup`/`Teardown` and devfs.
- GitHub Actions CI: gofmt, `go mod tidy`, Linux build and unit tests, and the
  full e2e suite as root in FreeBSD 14.5 and 15.1 VMs under QEMU.
- Community files: contributing guide, code of conduct, security policy,
  changelog, issue and PR templates, Dependabot configuration.

### Fixed

- `VNETInterface` always failed with "unknown parameter: vnet.interface":
  it is a jail(8) pseudo-parameter, not a kernel one. Interfaces are now moved
  with `SIOCSIFVNET` after the jail is created.
- `MountDevFS` always failed with "unknown parameter: mount.devfs" (same cause).
- An unknown jail parameter was reported as `ErrNotFound` ("jail not found");
  it is now `ErrInvalidArgument`.
- `CreateOrUpdate` on an existing VNET jail failed with "vnet cannot be
  changed after creation"; `vnet`, like `path`, is no longer sent on update.
- `vnet.Setup` failed with "create epair: invalid argument" when `if_epair`
  was not loaded (fixed in go-freebsd-ifc, which now loads it).
- `go build ./...` failed on non-FreeBSD systems because `internal/c/jail_wrap.c`
  had no build constraint.

## [1.0.0] - 2026-04-26

### Added

- Initial release: create, update, remove, get and list jails through libjail;
  command execution with `jexec`; VNET helpers; `jail.conf` rendering and
  parsing; typed errors; `ErrNotSupported` stubs on non-FreeBSD systems.

[Unreleased]: https://github.com/zombocoder/go-freebsd-jail/compare/v1.0.0...HEAD
[1.0.0]: https://github.com/zombocoder/go-freebsd-jail/releases/tag/v1.0.0
