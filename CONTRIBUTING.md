# Contributing to go-freebsd-jail

Thanks for your interest in improving the library! Bug reports, fixes, new
jail features, tests, examples and documentation are all welcome.

By participating you agree to follow the [Code of Conduct](CODE_OF_CONDUCT.md).
Security problems should be reported privately, see [SECURITY.md](SECURITY.md).

## Table of Contents

- [Getting Started](#getting-started)
- [Running Tests](#running-tests)
- [Writing Tests](#writing-tests)
- [Continuous Integration](#continuous-integration)
- [Code Style](#code-style)
- [Submitting Changes](#submitting-changes)
- [Architecture Overview](#architecture-overview)
- [Where Help Is Wanted](#where-help-is-wanted)

## Getting Started

### Prerequisites

- FreeBSD 14 or 15 with a `VIMAGE` kernel (the default `GENERIC`), amd64 or arm64
- Go 1.24+ (`pkg install go`)
- The base system C compiler and libjail (both in base)
- Root (or `doas`) for e2e tests and for examples that create jails

A FreeBSD virtual machine is the most comfortable setup: e2e tests create and
remove jails, vnet interfaces and devfs mounts.

The module also builds on other systems (every public function returns
`ErrNotSupported` there), so editing and running unit tests on Linux or macOS
works; anything that touches jails needs FreeBSD.

### Setup

```sh
git clone https://github.com/<your-username>/go-freebsd-jail.git
cd go-freebsd-jail
make build
make test
```

### Working on go-freebsd-ifc at the same time

Network interface handling (epair, bridge, moving interfaces into a vnet)
lives in [go-freebsd-ifc](https://github.com/zombocoder/go-freebsd-ifc). To
develop both together, clone it next to this repository and use a Go
workspace; `go.work` is git-ignored:

```sh
go work init . ../go-freebsd-ifc
```

## Running Tests

| Command | Runs as | What it covers |
|---|---|---|
| `make test` | any user, any OS | Unit tests in `tests/unit`: parameter conversion, validation, errors, `jail.conf` rendering and parsing. |
| `doas make test-e2e` | root, FreeBSD | E2E tests in `tests/e2e` (build tag `e2e`): jail lifecycle, `jexec`, VNET jails and `vnet.Setup`, devfs mounts. |

Running a subset works the usual way:

```sh
doas go test -count=1 -tags=e2e -run VNET -v ./tests/e2e/
```

E2E tests only touch jails they create (named `gotest-*-<pid>`), temporary
directories, and interfaces they create. VNET tests use `/` as the jail root
because they run `/bin/sh`, `ifconfig` and `route` inside the jail; devfs tests
use an empty temporary root. Prefer a VM or test machine over your workstation.

## Writing Tests

- **Test first.** Write a test that fails for the right reason, then the fix.
- **Check independently.** Where possible, verify results with the system
  (`statfs`/`mount` for devfs, `ifconfig` inside the jail via `jexec`, `jls`)
  rather than only with the library's own getters.
- **Test rollback.** For operations with several steps (create jail, mount
  devfs, move interfaces), add a test that makes a later step fail and checks
  that nothing is left behind.
- **Never** touch existing jails or host interfaces; register cleanup with
  `t.Cleanup` so jails and mounts are removed even when a test fails.

## Continuous Integration

[`.github/workflows/ci.yml`](.github/workflows/ci.yml) runs on every push to
`main` and every pull request:

- **Lint and non-FreeBSD build** (Linux): `gofmt`, `go mod tidy -diff`,
  `go vet`, build of all packages and examples, and the unit tests.
- **FreeBSD 14.5 and 15.1** (QEMU VMs via
  [vmactions/freebsd-vm](https://github.com/vmactions/freebsd-vm)): installs Go,
  runs `go vet` (also with the `e2e` tag), builds everything, and runs the unit
  tests and the full e2e suite as root.

## Code Style

### Go

- Standard Go conventions; `gofmt` is enforced by CI.
- Doc comments on all exported identifiers.
- Wrap errors in `OperationError` / `ValidationError` and keep the sentinel
  errors (`ErrNotFound`, `ErrExists`, `ErrInvalidArgument`, ...) reachable with
  `errors.Is`.
- Validate input before any system call.
- Every FreeBSD-only file has a `!freebsd` counterpart returning
  `ErrNotSupported`.

### Kernel parameters vs. jail(8)

`jail_set(2)` only accepts real kernel parameters (see `sysctl -d security.jail.param`).
Many familiar `jail.conf` settings — `vnet.interface`, `mount.devfs`,
`mount.fstab`, `exec.start`, `ip4.addr` with an interface prefix — are
implemented by jail(8) in userland. Supporting one of those means doing what
jail(8) does (see `usr.sbin/jail/command.c`) in `internal/jailops`, with
rollback on failure, never passing it to the kernel.

### cgo

- All cgo lives in `internal/c`; every C allocation is freed.
- libjail keeps its error message in a global (`jail_errmsg`); calls into it
  are serialized with the package mutex.

## Submitting Changes

1. Create a branch from `main`.
2. Make the change with tests; update doc comments, README and examples when
   the public API changes, and add an entry to [CHANGELOG.md](CHANGELOG.md)
   under "Unreleased".
3. Run `make check` and `doas make test-e2e` on FreeBSD.
4. Open a pull request; the template has a short checklist. CI must be green.

Commit messages: `type(scope): summary` (e.g. `fix(jailops): mount devfs
before creating the jail`), a blank line, then what changed and why. Common
types are `feat`, `fix`, `docs`, `test`, `refactor`, `ci`, `chore`.

### Scope

- The library manages jails directly through libjail and system calls.
  Shelling out is limited to `jexec` (see `exec/`) and tests.
- Avoid breaking public API changes within a major version. When one is
  unavoidable, call it out in the pull request and in the changelog.

## Architecture Overview

```
jail/           Public API: Create, CreateOrUpdate, Update, Remove, Get, List, Exists
exec/           Run commands inside a jail via jexec(8)
vnet/           VNET networking helpers (epair, bridge, addressing) on top of go-freebsd-ifc
jailconf/       Render and parse jail.conf
internal/
  c/            cgo bindings: libjail (jailparam_*), nmount/unmount
  jailops/      Operations: parameter handling, interface moves, devfs, rollback
  param/        Config <-> kernel parameter conversion
  validate/     Input validation
  jailerr/      Error types
  types/        Shared types (Config, Info, ...)
tests/unit/     Unit tests (any OS)
tests/e2e/      E2E tests (FreeBSD, root, build tag e2e)
examples/       Example programs
```

## Where Help Is Wanted

- **`mount.fstab` / nullfs mounts** with cleanup on removal
- **`exec.start` / `exec.stop`** lifecycle hooks
- **Resource limits** via `rctl(4)`
- **Jail attach in-process** for a forked child (safe `jail_attach` helper)
- **`jail.conf` round-trips** for more parameters and variables

Questions and design discussions are welcome in GitHub issues.
