# go-freebsd-jail

[![CI](https://github.com/zombocoder/go-freebsd-jail/actions/workflows/ci.yml/badge.svg)](https://github.com/zombocoder/go-freebsd-jail/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/zombocoder/go-freebsd-jail.svg)](https://pkg.go.dev/github.com/zombocoder/go-freebsd-jail)
[![Go Report Card](https://goreportcard.com/badge/github.com/zombocoder/go-freebsd-jail)](https://goreportcard.com/report/github.com/zombocoder/go-freebsd-jail)
[![Release](https://img.shields.io/github/v/release/zombocoder/go-freebsd-jail?sort=semver)](https://github.com/zombocoder/go-freebsd-jail/releases)
[![License: BSD-2-Clause](https://img.shields.io/badge/License-BSD--2--Clause-orange.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/Go-%3E%3D1.24-00ADD8.svg)](https://golang.org/)
[![FreeBSD](https://img.shields.io/badge/FreeBSD-14%20%7C%2015-red.svg)](https://www.freebsd.org/)

FreeBSD jail management library for Go with native cgo/libjail bindings.

**Tested in CI on:** FreeBSD 14.5 and 15.1 (amd64), with the full e2e suite
running as root in QEMU VMs. See the [changelog](CHANGELOG.md) and
[releases](https://github.com/zombocoder/go-freebsd-jail/releases).

## Requirements

- FreeBSD 14 or 15 (tested in CI on 14.5 and 15.1); VNET needs a `VIMAGE` kernel (the default `GENERIC`)
- Go 1.24+
- cgo enabled (FreeBSD base C compiler)
- libjail (included in FreeBSD base)
- Root privileges for create/update/remove/attach operations

## Installation

```bash
go get github.com/zombocoder/go-freebsd-jail
```

## Quick Start

### Create a Jail

```go
package main

import (
    "fmt"
    "log"

    "github.com/zombocoder/go-freebsd-jail/jail"
)

func main() {
    jid, err := jail.Create(jail.Config{
        Name:     "web",
        Path:     "/jails/web",
        Hostname: "web.local",
        Persist:  true,
    })
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Created jail with JID %d\n", jid)
}
```

### List Jails

```go
jails, err := jail.List()
if err != nil {
    log.Fatal(err)
}
for _, j := range jails {
    fmt.Printf("JID=%d Name=%s Hostname=%s Path=%s\n", j.JID, j.Name, j.Hostname, j.Path)
}
```

### Execute Commands

```go
import "github.com/zombocoder/go-freebsd-jail/exec"

out, err := exec.Output("web", "/bin/freebsd-version")
if err != nil {
    log.Fatal(err)
}
fmt.Println(string(out))
```

### Remove a Jail

```go
err := jail.Remove("web")
```

> **Warning:** `jail.Remove` kills all processes in the jail and removes child jails. This is a destructive operation.

### Create a VNET Jail

```go
jid, err := jail.Create(jail.Config{
    Name:          "web",
    Path:          "/jails/web",
    Hostname:      "web.local",
    Persist:       true,
    VNET:          true,
    VNETInterface: []string{"epair0b"}, // moved into the jail's vnet after creation
})
```

### Mount devfs

```go
jid, err := jail.Create(jail.Config{
    Name:       "web",
    Path:       "/jails/web",
    Persist:    true,
    MountDevFS: true, // devfs on /jails/web/dev with ruleset 4 (devfsrules_jail)
})
// jail.Remove("web") unmounts it again.
```

Like `jail(8)`, the library implements `vnet.interface` and `mount.devfs`
itself (they are not kernel parameters) and rolls back the jail if either step
fails.

### VNET Network Setup

```go
import "github.com/zombocoder/go-freebsd-jail/vnet"

netCfg := vnet.NetworkConfig{
    JailName:    "web",
    Bridge:      "bridge0",
    CreateEpair: true,
    JailAddress: "10.0.0.2/24",
    Gateway:     "10.0.0.1",
}
// Setup fills in netCfg.EpairHost/EpairJail; pass the same value to Teardown.
if err := vnet.Setup(&netCfg); err != nil {
    log.Fatal(err)
}
defer vnet.Teardown(&netCfg)
```

### Generate jail.conf

```go
import "github.com/zombocoder/go-freebsd-jail/jailconf"

conf, err := jailconf.Render("web", jail.Config{
    Path:     "/jails/web",
    Hostname: "web.local",
    Persist:  true,
    VNET:     true,
})
fmt.Println(conf)
```

## Error Handling

All errors support `errors.Is` and `errors.As`:

```go
jid, err := jail.Create(cfg)
if errors.Is(err, jail.ErrExists) {
    // jail already exists
}
if errors.Is(err, jail.ErrPermission) {
    // need root privileges
}

var opErr *jail.OperationError
if errors.As(err, &opErr) {
    fmt.Printf("operation %s on jail %s failed: %v\n", opErr.Op, opErr.Jail, opErr.Err)
}
```

## Root Privileges

Most jail operations require root privileges. Operations that fail due to insufficient permissions return `ErrPermission`.

## FreeBSD Only

This library only works on FreeBSD. On other operating systems, all public functions return `ErrNotSupported`. The library compiles cleanly on all platforms — no cgo is required on non-FreeBSD systems.

## Contributing

Contributions are welcome — see [CONTRIBUTING.md](CONTRIBUTING.md) for how to
run the tests, what CI checks, and where help is wanted. Please follow the
[Code of Conduct](CODE_OF_CONDUCT.md) and report security issues privately as
described in [SECURITY.md](SECURITY.md).

## Related Projects

- [go-freebsd-ifc](https://github.com/zombocoder/go-freebsd-ifc) — FreeBSD network interface control library (used by the `vnet` package)

## License

BSD-2-Clause
