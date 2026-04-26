package unit

import (
	"testing"

	"github.com/zombocoder/go-freebsd-jail/internal/param"
	"github.com/zombocoder/go-freebsd-jail/internal/types"
)

func TestConfigToParams_Basic(t *testing.T) {
	cfg := types.Config{
		Name:     "web",
		Path:     "/jails/web",
		Hostname: "web.local",
		Persist:  true,
	}

	params, err := param.ConfigToParams(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := map[string]struct {
		value  string
		isBool bool
	}{
		"name":          {value: "web"},
		"path":          {value: "/jails/web"},
		"host.hostname": {value: "web.local"},
		"persist":       {isBool: true},
	}

	if len(params) != len(expected) {
		t.Fatalf("expected %d params, got %d", len(expected), len(params))
	}

	for _, p := range params {
		exp, ok := expected[p.Name]
		if !ok {
			t.Errorf("unexpected param %q", p.Name)
			continue
		}
		if p.Value != exp.value {
			t.Errorf("param %q: expected value %q, got %q", p.Name, exp.value, p.Value)
		}
		if p.IsBool != exp.isBool {
			t.Errorf("param %q: expected isBool=%v, got %v", p.Name, exp.isBool, p.IsBool)
		}
	}
}

func TestConfigToParams_VNET(t *testing.T) {
	cfg := types.Config{
		Name:          "vnet-jail",
		Path:          "/jails/vnet",
		Hostname:      "vnet.local",
		Persist:       true,
		VNET:          true,
		VNETInterface: []string{"epair0b"},
	}

	params, err := param.ConfigToParams(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	found := make(map[string]bool)
	for _, p := range params {
		found[p.Name] = true
	}

	for _, required := range []string{"name", "path", "host.hostname", "persist", "vnet", "vnet.interface"} {
		if !found[required] {
			t.Errorf("missing required param %q", required)
		}
	}
}

func TestConfigToParams_AllowConfig(t *testing.T) {
	cfg := types.Config{
		Name:     "allow-test",
		Path:     "/jails/allow",
		Hostname: "allow.local",
		Allow: types.AllowConfig{
			RawSockets:  true,
			Mount:       true,
			SysVIPC:     true,
			SetHostname: true,
		},
	}

	params, err := param.ConfigToParams(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	found := make(map[string]bool)
	for _, p := range params {
		found[p.Name] = true
		if p.Name == "allow.raw_sockets" || p.Name == "allow.mount" ||
			p.Name == "allow.sysvipc" || p.Name == "allow.set_hostname" {
			if !p.IsBool {
				t.Errorf("param %q should be bool", p.Name)
			}
		}
	}

	for _, required := range []string{"allow.raw_sockets", "allow.mount", "allow.sysvipc", "allow.set_hostname"} {
		if !found[required] {
			t.Errorf("missing allow param %q", required)
		}
	}
}

func TestConfigToParams_IP(t *testing.T) {
	cfg := types.Config{
		Name:     "ip-test",
		Path:     "/jails/ip",
		Hostname: "ip.local",
		IP4:      []string{"192.168.1.1", "10.0.0.1"},
		IP6:      []string{"::1"},
	}

	params, err := param.ConfigToParams(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, p := range params {
		if p.Name == "ip4.addr" {
			if p.Value != "192.168.1.1,10.0.0.1" {
				t.Errorf("ip4.addr: expected %q, got %q", "192.168.1.1,10.0.0.1", p.Value)
			}
		}
		if p.Name == "ip6.addr" {
			if p.Value != "::1" {
				t.Errorf("ip6.addr: expected %q, got %q", "::1", p.Value)
			}
		}
	}
}

func TestConfigToParams_ReservedRawParam(t *testing.T) {
	cfg := types.Config{
		Name:      "test",
		Path:      "/jails/test",
		Hostname:  "test.local",
		RawParams: map[string]string{"name": "override"},
	}

	_, err := param.ConfigToParams(cfg)
	if err == nil {
		t.Fatal("expected error for reserved RawParam, got nil")
	}
}

func TestConfigToParams_Validation(t *testing.T) {
	tests := []struct {
		name string
		cfg  types.Config
	}{
		{"invalid name", types.Config{Name: "1bad", Path: "/jails/x", Hostname: "x"}},
		{"relative path", types.Config{Name: "ok", Path: "relative", Hostname: "x"}},
		{"invalid hostname", types.Config{Name: "ok", Path: "/jails/x", Hostname: "bad..host"}},
		{"invalid ip", types.Config{Name: "ok", Path: "/jails/x", Hostname: "x", IP4: []string{"not-ip"}}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := param.ConfigToParams(tt.cfg)
			if err == nil {
				t.Error("expected validation error, got nil")
			}
		})
	}
}

func TestValuesToInfo(t *testing.T) {
	values := map[string]string{
		"name":          "web",
		"path":          "/jails/web",
		"host.hostname": "web.local",
		"ip4.addr":      "192.168.1.1,10.0.0.1",
		"ip6.addr":      "::1",
		"vnet":          "new",
		"dying":         "0",
	}

	info := param.ValuesToInfo(42, values)

	if info.JID != 42 {
		t.Errorf("JID: expected 42, got %d", info.JID)
	}
	if info.Name != "web" {
		t.Errorf("Name: expected %q, got %q", "web", info.Name)
	}
	if info.Path != "/jails/web" {
		t.Errorf("Path: expected %q, got %q", "/jails/web", info.Path)
	}
	if info.Hostname != "web.local" {
		t.Errorf("Hostname: expected %q, got %q", "web.local", info.Hostname)
	}
	if len(info.IP4) != 2 {
		t.Errorf("IP4: expected 2 entries, got %d", len(info.IP4))
	}
	if !info.VNET {
		t.Error("VNET: expected true")
	}
	if info.State != types.StateActive {
		t.Errorf("State: expected %q, got %q", types.StateActive, info.State)
	}
}
