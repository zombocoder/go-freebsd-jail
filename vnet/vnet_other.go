//go:build !freebsd

package vnet

import "github.com/zombocoder/go-freebsd-jail/jail"

// NetworkConfig describes the VNET networking setup for a jail.
type NetworkConfig struct {
	JailName          string
	Bridge            string
	CreateBridge      bool
	EpairHost         string
	EpairJail         string
	CreateEpair       bool
	JailInterfaceName string
	HostAddress       string
	JailAddress       string
	Gateway           string
}

// Setup configures VNET networking. Not supported on this platform.
func Setup(cfg NetworkConfig) error {
	return jail.ErrNotSupported
}

// Teardown removes VNET networking. Not supported on this platform.
func Teardown(cfg NetworkConfig) error {
	return jail.ErrNotSupported
}
