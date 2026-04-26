package unit

import (
	"strings"
	"testing"

	"github.com/zombocoder/go-freebsd-jail/jail"
	"github.com/zombocoder/go-freebsd-jail/jailconf"
)

func TestRender_Basic(t *testing.T) {
	cfg := jail.Config{
		Path:     "/jails/web",
		Hostname: "web.local",
		Persist:  true,
	}

	out, err := jailconf.Render("web", cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	mustContain(t, out, `web {`)
	mustContain(t, out, `path = "/jails/web";`)
	mustContain(t, out, `host.hostname = "web.local";`)
	mustContain(t, out, `persist;`)
	mustContain(t, out, `}`)
}

func TestRender_VNET(t *testing.T) {
	cfg := jail.Config{
		Path:          "/jails/vnet",
		Hostname:      "vnet.local",
		Persist:       true,
		VNET:          true,
		VNETInterface: []string{"epair0b"},
		MountDevFS:    true,
		Allow: jail.AllowConfig{
			RawSockets: true,
		},
	}

	out, err := jailconf.Render("vnet-jail", cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	mustContain(t, out, `vnet;`)
	mustContain(t, out, `vnet.interface = "epair0b";`)
	mustContain(t, out, `mount.devfs;`)
	mustContain(t, out, `allow.raw_sockets;`)
}

func TestRender_IP(t *testing.T) {
	cfg := jail.Config{
		Path:     "/jails/ip",
		Hostname: "ip.local",
		IP4:      []string{"192.168.1.1", "10.0.0.1"},
	}

	out, err := jailconf.Render("ip-jail", cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	mustContain(t, out, `ip4.addr = "192.168.1.1, 10.0.0.1";`)
}

func TestRenderMany(t *testing.T) {
	configs := map[string]jail.Config{
		"a": {Path: "/jails/a", Hostname: "a.local", Persist: true},
		"b": {Path: "/jails/b", Hostname: "b.local"},
	}

	out, err := jailconf.RenderMany(configs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	mustContain(t, out, `a {`)
	mustContain(t, out, `b {`)

	// Sorted: a before b
	aIdx := strings.Index(out, "a {")
	bIdx := strings.Index(out, "b {")
	if aIdx >= bIdx {
		t.Error("expected jail 'a' before jail 'b' in output")
	}
}

func TestParse_Basic(t *testing.T) {
	input := `web {
    path = "/jails/web";
    host.hostname = "web.local";
    persist;
    vnet;
    vnet.interface = "epair0b";
    mount.devfs;
    allow.raw_sockets;
}
`

	configs, err := jailconf.Parse([]byte(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	cfg, ok := configs["web"]
	if !ok {
		t.Fatal("expected jail 'web' in results")
	}

	if cfg.Path != "/jails/web" {
		t.Errorf("Path: expected %q, got %q", "/jails/web", cfg.Path)
	}
	if cfg.Hostname != "web.local" {
		t.Errorf("Hostname: expected %q, got %q", "web.local", cfg.Hostname)
	}
	if !cfg.Persist {
		t.Error("expected Persist=true")
	}
	if !cfg.VNET {
		t.Error("expected VNET=true")
	}
	if len(cfg.VNETInterface) != 1 || cfg.VNETInterface[0] != "epair0b" {
		t.Errorf("VNETInterface: expected [epair0b], got %v", cfg.VNETInterface)
	}
	if !cfg.MountDevFS {
		t.Error("expected MountDevFS=true")
	}
	if !cfg.Allow.RawSockets {
		t.Error("expected Allow.RawSockets=true")
	}
}

func TestParse_Multiple(t *testing.T) {
	input := `# Test config
web {
    path = "/jails/web";
    host.hostname = "web.local";
    persist;
}

db {
    path = "/jails/db";
    host.hostname = "db.local";
    allow.sysvipc;
}
`

	configs, err := jailconf.Parse([]byte(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(configs) != 2 {
		t.Fatalf("expected 2 jails, got %d", len(configs))
	}

	if _, ok := configs["web"]; !ok {
		t.Error("expected jail 'web'")
	}
	if _, ok := configs["db"]; !ok {
		t.Error("expected jail 'db'")
	}

	if !configs["db"].Allow.SysVIPC {
		t.Error("expected db.Allow.SysVIPC=true")
	}
}

func TestParse_RoundTrip(t *testing.T) {
	original := jail.Config{
		Path:     "/jails/roundtrip",
		Hostname: "rt.local",
		Persist:  true,
		VNET:     true,
		VNETInterface: []string{"epair0b"},
		MountDevFS: true,
		Allow: jail.AllowConfig{
			RawSockets:  true,
			Mount:       true,
			SetHostname: true,
		},
	}

	rendered, err := jailconf.Render("roundtrip", original)
	if err != nil {
		t.Fatalf("render error: %v", err)
	}

	parsed, err := jailconf.Parse([]byte(rendered))
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	cfg, ok := parsed["roundtrip"]
	if !ok {
		t.Fatal("expected jail 'roundtrip' in parsed output")
	}

	if cfg.Path != original.Path {
		t.Errorf("Path: expected %q, got %q", original.Path, cfg.Path)
	}
	if cfg.Hostname != original.Hostname {
		t.Errorf("Hostname: expected %q, got %q", original.Hostname, cfg.Hostname)
	}
	if cfg.Persist != original.Persist {
		t.Errorf("Persist: expected %v, got %v", original.Persist, cfg.Persist)
	}
	if cfg.VNET != original.VNET {
		t.Errorf("VNET: expected %v, got %v", original.VNET, cfg.VNET)
	}
	if cfg.Allow.RawSockets != original.Allow.RawSockets {
		t.Errorf("Allow.RawSockets: expected %v, got %v", original.Allow.RawSockets, cfg.Allow.RawSockets)
	}
}

func TestParse_Comments(t *testing.T) {
	input := `# This is a comment
# Another comment
web {
    # Path to jail
    path = "/jails/web";
    host.hostname = "web.local";
}
`

	configs, err := jailconf.Parse([]byte(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	cfg, ok := configs["web"]
	if !ok {
		t.Fatal("expected jail 'web'")
	}
	if cfg.Path != "/jails/web" {
		t.Errorf("Path: expected %q, got %q", "/jails/web", cfg.Path)
	}
}

func mustContain(t *testing.T, s, substr string) {
	t.Helper()
	if !strings.Contains(s, substr) {
		t.Errorf("expected output to contain %q, got:\n%s", substr, s)
	}
}
