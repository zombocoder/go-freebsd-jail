# go-freebsd-jail

[![Go Reference](https://pkg.go.dev/badge/github.com/zombocoder/go-freebsd-jail.svg)](https://pkg.go.dev/github.com/zombocoder/go-freebsd-jail)
[![FreeBSD](https://img.shields.io/badge/platform-FreeBSD-red.svg)](https://www.freebsd.org/)
[![Go Version](https://img.shields.io/badge/go-%3E%3D1.22-blue.svg)](https://golang.org/)
[![License](https://img.shields.io/badge/license-BSD--2--Clause-green.svg)](LICENSE)

FreeBSD jail management library for Go with native cgo/libjail bindings.

## Requirements

- FreeBSD 13.x+ (14.x primary target)
- Go 1.22+
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
    VNETInterface: []string{"epair0b"},
})
```

### VNET Network Setup

```go
import "github.com/zombocoder/go-freebsd-jail/vnet"

err := vnet.Setup(vnet.NetworkConfig{
    JailName:    "web",
    Bridge:      "bridge0",
    CreateEpair: true,
    JailAddress: "10.0.0.2/24",
    Gateway:     "10.0.0.1",
})
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

## Related Projects

- [go-freebsd-ifc](https://github.com/zombocoder/go-freebsd-ifc) — FreeBSD network interface control library (used by the `vnet` package)

## License

BSD-2-Clause
