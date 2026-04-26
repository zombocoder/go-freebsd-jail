# Features

## Jail Lifecycle Management
- Create, update, and remove FreeBSD jails programmatically
- Idempotent `CreateOrUpdate` for declarative jail management
- Query jail details and list all running jails
- Full support for jail parameters (persist, VNET, allow.*, limits)

## Native C/cgo Bindings
- Direct integration with FreeBSD libjail (`jailparam_*` APIs)
- Uses `jail_set(2)`, `jail_get(2)`, `jail_remove(2)` system calls
- Proper error mapping from errno to typed Go errors

## Command Execution
- Execute commands inside jails via `jexec`
- Capture stdout/stderr output
- Shell command support
- Safe subprocess model (no in-process jail_attach)

## VNET Networking
- Higher-level helpers for VNET jail network setup
- Integrates with [go-freebsd-ifc](https://github.com/zombocoder/go-freebsd-ifc) for epair/bridge management
- Automatic epair creation, bridge membership, and in-jail IP/route configuration
- Idempotent setup and teardown

## jail.conf Generation and Parsing
- Render `jail.conf`-compatible configuration from Go structs
- Parse existing `jail.conf` files back into Go structs
- Multi-jail rendering support

## Cross-Platform Safety
- All public functions return `ErrNotSupported` on non-FreeBSD systems
- Build tags ensure clean compilation on any GOOS
- No cgo required on non-FreeBSD platforms

## Error Handling
- Typed sentinel errors (`ErrNotFound`, `ErrPermission`, `ErrExists`, etc.)
- Structured `OperationError` and `ValidationError` types
- Full `errors.Is` / `errors.As` support
- Input validation before any system calls
