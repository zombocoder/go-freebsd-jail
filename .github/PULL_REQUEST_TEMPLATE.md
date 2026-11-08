## What and why

<!-- What does this change do, and what problem does it solve? Link issues with "Fixes #123". -->

## How it was tested

<!-- CI runs everything on FreeBSD 14.x and 15.x VMs and builds on Linux. If you also tested locally, say how, e.g.:
     doas go test -count=1 -tags=e2e ./tests/e2e/... on FreeBSD 15.1 -->

## Checklist

- [ ] New behavior is covered by tests (a failing test first, then the fix)
- [ ] E2E tests use unique jail names (`gotest-*-<pid>`), never touch existing jails, and clean up with `t.Cleanup`
- [ ] Where possible, tests check the result independently (`jls`, `mount`, `ifconfig`) rather than only with the library itself
- [ ] Only real kernel parameters reach `jail_set(2)`; jail(8) pseudo-parameters (`vnet.interface`, `mount.*`, `exec.*`) are implemented explicitly
- [ ] cgo code frees every C allocation; non-FreeBSD stubs return `ErrNotSupported`
- [ ] Public API changes are documented (doc comments, README, examples) and listed in `CHANGELOG.md`
- [ ] `gofmt -l .` prints nothing
