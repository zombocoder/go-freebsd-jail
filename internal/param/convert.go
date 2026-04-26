package param

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/zombocoder/go-freebsd-jail/internal/jailerr"
	"github.com/zombocoder/go-freebsd-jail/internal/types"
	"github.com/zombocoder/go-freebsd-jail/internal/validate"
)

// reservedParams are params managed by Config fields and cannot be overridden via RawParams.
var reservedParams = map[string]bool{
	"name": true, "path": true, "host.hostname": true,
	"persist": true, "nopersist": true,
	"vnet": true, "novnet": true,
	"vnet.interface": true,
	"ip4.addr": true, "ip6.addr": true,
	"children.max": true, "devfs_ruleset": true,
	"jid": true, "lastjid": true,
}

// ConfigToParams converts a Config to a list of Params for the C layer.
func ConfigToParams(cfg types.Config) ([]Param, error) {
	var params []Param

	if cfg.Name != "" {
		if err := validate.JailName(cfg.Name); err != nil {
			return nil, err
		}
		params = append(params, Param{Name: "name", Value: cfg.Name})
	}

	if cfg.Path != "" {
		if err := validate.Path(cfg.Path); err != nil {
			return nil, err
		}
		params = append(params, Param{Name: "path", Value: cfg.Path})
	}

	if cfg.Hostname != "" {
		if err := validate.Hostname(cfg.Hostname); err != nil {
			return nil, err
		}
		params = append(params, Param{Name: "host.hostname", Value: cfg.Hostname})
	}

	if cfg.Persist {
		params = append(params, Param{Name: "persist", IsBool: true})
	}

	if cfg.VNET {
		params = append(params, Param{Name: "vnet", IsBool: true})
	}

	for _, iface := range cfg.VNETInterface {
		if err := validate.InterfaceName(iface); err != nil {
			return nil, err
		}
		params = append(params, Param{Name: "vnet.interface", Value: iface})
	}

	if len(cfg.IP4) > 0 {
		for _, addr := range cfg.IP4 {
			if err := validate.IPAddress(addr); err != nil {
				return nil, err
			}
		}
		params = append(params, Param{Name: "ip4.addr", Value: strings.Join(cfg.IP4, ",")})
	}

	if len(cfg.IP6) > 0 {
		for _, addr := range cfg.IP6 {
			if err := validate.IPAddress(addr); err != nil {
				return nil, err
			}
		}
		params = append(params, Param{Name: "ip6.addr", Value: strings.Join(cfg.IP6, ",")})
	}

	if cfg.MountDevFS {
		params = append(params, Param{Name: "mount.devfs", IsBool: true})
	}

	if cfg.DevFSRuleset != nil {
		params = append(params, Param{Name: "devfs_ruleset", Value: strconv.Itoa(*cfg.DevFSRuleset)})
	}

	params = appendAllowParams(params, cfg.Allow)

	if cfg.Limits.ChildrenMax != nil {
		params = append(params, Param{Name: "children.max", Value: strconv.Itoa(*cfg.Limits.ChildrenMax)})
	}

	for k, v := range cfg.RawParams {
		if reservedParams[k] {
			return nil, &jailerr.ValidationError{
				Field: "RawParams",
				Value: k,
				Err:   fmt.Errorf("parameter %q is reserved and cannot be set via RawParams", k),
			}
		}
		if v == "" {
			params = append(params, Param{Name: k, IsBool: true})
		} else {
			params = append(params, Param{Name: k, Value: v})
		}
	}

	return params, nil
}

func appendAllowParams(params []Param, allow types.AllowConfig) []Param {
	if allow.RawSockets {
		params = append(params, Param{Name: "allow.raw_sockets", IsBool: true})
	}
	if allow.Mount {
		params = append(params, Param{Name: "allow.mount", IsBool: true})
	}
	if allow.MountDevFS {
		params = append(params, Param{Name: "allow.mount.devfs", IsBool: true})
	}
	if allow.MountNullFS {
		params = append(params, Param{Name: "allow.mount.nullfs", IsBool: true})
	}
	if allow.MountProcFS {
		params = append(params, Param{Name: "allow.mount.procfs", IsBool: true})
	}
	if allow.MountTmpFS {
		params = append(params, Param{Name: "allow.mount.tmpfs", IsBool: true})
	}
	if allow.SysVIPC {
		params = append(params, Param{Name: "allow.sysvipc", IsBool: true})
	}
	if allow.SetHostname {
		params = append(params, Param{Name: "allow.set_hostname", IsBool: true})
	}
	if allow.Chflags {
		params = append(params, Param{Name: "allow.chflags", IsBool: true})
	}
	if allow.SocketAF {
		params = append(params, Param{Name: "allow.socket_af", IsBool: true})
	}
	return params
}

// ValuesToInfo converts a map of param name->value from JailGet into an Info struct.
func ValuesToInfo(jid int, values map[string]string) *types.Info {
	info := &types.Info{
		JID:      types.JID(jid),
		Name:     values["name"],
		Path:     values["path"],
		Hostname: values["host.hostname"],
		Params:   values,
	}

	if ip4 := values["ip4.addr"]; ip4 != "" {
		info.IP4 = strings.Split(ip4, ",")
	}
	if ip6 := values["ip6.addr"]; ip6 != "" {
		info.IP6 = strings.Split(ip6, ",")
	}

	if vnet := values["vnet"]; vnet == "new" || vnet == "inherit" {
		info.VNET = vnet == "new"
	}

	dying := values["dying"]
	if dying == "1" {
		info.State = types.StateDying
	} else {
		info.State = types.StateActive
	}

	return info
}

// StandardGetParams returns the list of standard parameter names to fetch with JailGet.
func StandardGetParams() []string {
	return []string{
		"name",
		"path",
		"host.hostname",
		"ip4.addr",
		"ip6.addr",
		"vnet",
		"dying",
		"persist",
		"children.max",
		"devfs_ruleset",
		"allow.raw_sockets",
		"allow.mount",
		"allow.sysvipc",
		"allow.set_hostname",
		"allow.chflags",
		"allow.socket_af",
		"enforce_statfs",
		"securelevel",
	}
}
