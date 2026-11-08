# Security Policy

## Supported versions

Security fixes are made on the `main` branch and released as a new patch or
minor version of the latest major release (currently 1.x).

## Reporting a vulnerability

Please **do not open a public issue** for security problems.

Report them privately through GitHub:
[Report a vulnerability](https://github.com/zombocoder/go-freebsd-jail/security/advisories/new)
(Security tab → "Report a vulnerability").

Include what you can of:

- the affected function or package and the commit you tested;
- the FreeBSD version (`freebsd-version -ku`) and architecture;
- steps or a minimal program that reproduces the problem;
- the impact you expect (jail escape, privilege issue, wrong jail or host
  resource modified, memory corruption, ...).

You should get a first response within a week. Once a fix is available, the
advisory is published with credit to the reporter unless you prefer otherwise.

## Scope

This library creates and manages jails through libjail via cgo and is used by
privileged processes, so jail isolation is what matters most. Of particular
interest are:

- configurations the library produces that are less isolated than requested,
  e.g. a devfs mounted without the expected ruleset, `allow.*` permissions
  that were not asked for, or interfaces left in the host vnet;
- operations that act on a different jail, path, mount or interface than
  requested (for example unmounting the host's `/dev`);
- input that reaches `jail_set(2)`, `nmount(2)` or `jexec(8)` without the
  validation callers would expect (names, paths, parameters, commands);
- memory-safety bugs in the cgo layer (`internal/c`).

Bugs in the FreeBSD jail subsystem itself should be reported to the
[FreeBSD security team](https://www.freebsd.org/security/).
