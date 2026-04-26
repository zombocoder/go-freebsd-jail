package validate

import (
	"fmt"
	"net"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/zombocoder/go-freebsd-jail/internal/jailerr"
)

const (
	maxJailNameLen = 249
	maxIfNameLen   = 15 // IFNAMSIZ - 1
)

// JailName validates a jail name.
func JailName(name string) error {
	if name == "" {
		return &jailerr.ValidationError{Field: "name", Value: name, Err: fmt.Errorf("must not be empty")}
	}
	if len(name) > maxJailNameLen {
		return &jailerr.ValidationError{Field: "name", Value: name, Err: fmt.Errorf("must not exceed %d characters", maxJailNameLen)}
	}
	if !isAlpha(rune(name[0])) {
		return &jailerr.ValidationError{Field: "name", Value: name, Err: fmt.Errorf("must start with a letter")}
	}
	for _, r := range name {
		if !isAlpha(r) && !unicode.IsDigit(r) && r != '-' && r != '_' && r != '.' {
			return &jailerr.ValidationError{Field: "name", Value: name, Err: fmt.Errorf("contains invalid character %q", r)}
		}
	}
	return nil
}

// Path validates a jail root path.
func Path(path string) error {
	if path == "" {
		return &jailerr.ValidationError{Field: "path", Value: path, Err: fmt.Errorf("must not be empty")}
	}
	if !filepath.IsAbs(path) {
		return &jailerr.ValidationError{Field: "path", Value: path, Err: fmt.Errorf("must be an absolute path")}
	}
	if strings.ContainsRune(path, 0) {
		return &jailerr.ValidationError{Field: "path", Value: path, Err: fmt.Errorf("must not contain null bytes")}
	}
	return nil
}

// Hostname validates a jail hostname.
func Hostname(host string) error {
	if host == "" {
		return &jailerr.ValidationError{Field: "hostname", Value: host, Err: fmt.Errorf("must not be empty")}
	}
	if len(host) > 255 {
		return &jailerr.ValidationError{Field: "hostname", Value: host, Err: fmt.Errorf("must not exceed 255 characters")}
	}
	for _, label := range strings.Split(host, ".") {
		if label == "" {
			return &jailerr.ValidationError{Field: "hostname", Value: host, Err: fmt.Errorf("contains empty label")}
		}
		if len(label) > 63 {
			return &jailerr.ValidationError{Field: "hostname", Value: host, Err: fmt.Errorf("label exceeds 63 characters")}
		}
		for _, r := range label {
			if !isAlpha(r) && !unicode.IsDigit(r) && r != '-' {
				return &jailerr.ValidationError{Field: "hostname", Value: host, Err: fmt.Errorf("contains invalid character %q", r)}
			}
		}
	}
	return nil
}

// IPAddress validates an IP address string (IPv4 or IPv6).
func IPAddress(addr string) error {
	// Strip interface|addr format (e.g. "em0|192.168.1.1/24")
	if idx := strings.Index(addr, "|"); idx >= 0 {
		addr = addr[idx+1:]
	}
	// Strip CIDR mask if present
	host := addr
	if idx := strings.Index(addr, "/"); idx >= 0 {
		host = addr[:idx]
	}
	if net.ParseIP(host) == nil {
		return &jailerr.ValidationError{Field: "ip", Value: addr, Err: fmt.Errorf("invalid IP address")}
	}
	return nil
}

// InterfaceName validates a FreeBSD network interface name.
func InterfaceName(name string) error {
	if name == "" {
		return &jailerr.ValidationError{Field: "interface", Value: name, Err: fmt.Errorf("must not be empty")}
	}
	if len(name) > maxIfNameLen {
		return &jailerr.ValidationError{Field: "interface", Value: name, Err: fmt.Errorf("must not exceed %d characters", maxIfNameLen)}
	}
	for _, r := range name {
		if !isAlpha(r) && !unicode.IsDigit(r) {
			return &jailerr.ValidationError{Field: "interface", Value: name, Err: fmt.Errorf("contains invalid character %q", r)}
		}
	}
	return nil
}

func isAlpha(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z')
}
