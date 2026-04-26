package unit

import (
	"errors"
	"testing"

	"github.com/zombocoder/go-freebsd-jail/internal/jailerr"
	"github.com/zombocoder/go-freebsd-jail/internal/validate"
)

func TestJailName(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid simple", "web", false},
		{"valid with dots", "web.local", false},
		{"valid with hyphens", "my-jail", false},
		{"valid with underscores", "my_jail", false},
		{"valid alphanumeric", "jail01", false},
		{"empty", "", true},
		{"starts with number", "1jail", true},
		{"starts with hyphen", "-jail", true},
		{"contains space", "my jail", true},
		{"contains slash", "my/jail", true},
		{"too long", string(make([]byte, 250)), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validate.JailName(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("JailName(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
			if err != nil {
				var ve *jailerr.ValidationError
				if !errors.As(err, &ve) {
					t.Errorf("expected ValidationError, got %T", err)
				}
			}
		})
	}
}

func TestPath(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid absolute", "/jails/web", false},
		{"valid root", "/", false},
		{"empty", "", true},
		{"relative", "jails/web", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validate.Path(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Path(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}

func TestHostname(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid simple", "web", false},
		{"valid fqdn", "web.example.com", false},
		{"empty", "", true},
		{"empty label", "web..com", true},
		{"label too long", string(make([]byte, 64)) + ".com", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validate.Hostname(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Hostname(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}

func TestIPAddress(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid ipv4", "192.168.1.1", false},
		{"valid ipv4 with cidr", "192.168.1.1/24", false},
		{"valid ipv4 with iface", "em0|192.168.1.1/24", false},
		{"valid ipv6", "::1", false},
		{"valid ipv6 full", "2001:db8::1", false},
		{"invalid", "not-an-ip", true},
		{"empty", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validate.IPAddress(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("IPAddress(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}

func TestInterfaceName(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid", "epair0b", false},
		{"valid bridge", "bridge0", false},
		{"empty", "", true},
		{"too long", "abcdefghijklmnop", true},
		{"invalid char", "ep-0", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validate.InterfaceName(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("InterfaceName(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}
