package jailconf

import (
	"fmt"
	"sort"
	"strings"

	"github.com/zombocoder/go-freebsd-jail/jail"
)

// Render generates a jail.conf block for a single jail.
func Render(name string, cfg jail.Config) (string, error) {
	if name == "" {
		return "", fmt.Errorf("jail name must not be empty")
	}

	var b strings.Builder
	b.WriteString(name)
	b.WriteString(" {\n")

	if err := renderParams(&b, cfg); err != nil {
		return "", err
	}

	b.WriteString("}\n")
	return b.String(), nil
}

// RenderMany generates jail.conf content for multiple jails.
// Jails are rendered in sorted order by name for deterministic output.
func RenderMany(configs map[string]jail.Config) (string, error) {
	names := make([]string, 0, len(configs))
	for name := range configs {
		names = append(names, name)
	}
	sort.Strings(names)

	var b strings.Builder
	for i, name := range names {
		if i > 0 {
			b.WriteString("\n")
		}
		block, err := Render(name, configs[name])
		if err != nil {
			return "", err
		}
		b.WriteString(block)
	}
	return b.String(), nil
}

func renderParams(b *strings.Builder, cfg jail.Config) error {
	if cfg.Path != "" {
		writeKV(b, "path", cfg.Path)
	}
	if cfg.Hostname != "" {
		writeKV(b, "host.hostname", cfg.Hostname)
	}
	if cfg.Persist {
		writeBool(b, "persist")
	}
	if cfg.VNET {
		writeBool(b, "vnet")
	}
	for _, iface := range cfg.VNETInterface {
		writeKV(b, "vnet.interface", iface)
	}
	if len(cfg.IP4) > 0 {
		writeKV(b, "ip4.addr", strings.Join(cfg.IP4, ", "))
	}
	if len(cfg.IP6) > 0 {
		writeKV(b, "ip6.addr", strings.Join(cfg.IP6, ", "))
	}
	if cfg.MountDevFS {
		writeBool(b, "mount.devfs")
	}
	if cfg.DevFSRuleset != nil {
		writeKV(b, "devfs_ruleset", fmt.Sprintf("%d", *cfg.DevFSRuleset))
	}

	renderAllow(b, cfg.Allow)

	if cfg.Limits.ChildrenMax != nil {
		writeKV(b, "children.max", fmt.Sprintf("%d", *cfg.Limits.ChildrenMax))
	}

	// Raw params in sorted order
	if len(cfg.RawParams) > 0 {
		keys := make([]string, 0, len(cfg.RawParams))
		for k := range cfg.RawParams {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			v := cfg.RawParams[k]
			if v == "" {
				writeBool(b, k)
			} else {
				writeKV(b, k, v)
			}
		}
	}

	return nil
}

func renderAllow(b *strings.Builder, allow jail.AllowConfig) {
	if allow.RawSockets {
		writeBool(b, "allow.raw_sockets")
	}
	if allow.Mount {
		writeBool(b, "allow.mount")
	}
	if allow.MountDevFS {
		writeBool(b, "allow.mount.devfs")
	}
	if allow.MountNullFS {
		writeBool(b, "allow.mount.nullfs")
	}
	if allow.MountProcFS {
		writeBool(b, "allow.mount.procfs")
	}
	if allow.MountTmpFS {
		writeBool(b, "allow.mount.tmpfs")
	}
	if allow.SysVIPC {
		writeBool(b, "allow.sysvipc")
	}
	if allow.SetHostname {
		writeBool(b, "allow.set_hostname")
	}
	if allow.Chflags {
		writeBool(b, "allow.chflags")
	}
	if allow.SocketAF {
		writeBool(b, "allow.socket_af")
	}
}

func writeKV(b *strings.Builder, key, value string) {
	fmt.Fprintf(b, "    %s = \"%s\";\n", key, value)
}

func writeBool(b *strings.Builder, key string) {
	fmt.Fprintf(b, "    %s;\n", key)
}
