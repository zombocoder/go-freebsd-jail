package main

// create-vnet creates a VNET jail, wires it to a new bridge through an epair,
// configures addressing inside the jail, prints the jail's view of the
// interface and tears everything down again.
//
// The jail root (/jails/example-vnet) must contain /bin/sh, /sbin/ifconfig
// and /sbin/route, since networking inside the jail is configured with them.

import (
	"fmt"
	"log"

	jailexec "github.com/zombocoder/go-freebsd-jail/exec"
	"github.com/zombocoder/go-freebsd-jail/jail"
	"github.com/zombocoder/go-freebsd-jail/vnet"
)

const jailName = "example-vnet"

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	jid, err := jail.Create(jail.Config{
		Name:     jailName,
		Path:     "/jails/example-vnet",
		Hostname: "example-vnet.local",
		Persist:  true,
		VNET:     true,
	})
	if err != nil {
		return fmt.Errorf("create jail: %w", err)
	}
	fmt.Printf("Created VNET jail with JID %d\n", jid)
	defer func() {
		if err := jail.Remove(jailName); err != nil {
			log.Printf("remove jail: %v", err)
			return
		}
		fmt.Println("Jail removed")
	}()

	netCfg := vnet.NetworkConfig{
		JailName:          jailName,
		CreateBridge:      true,
		CreateEpair:       true,
		HostAddress:       "10.0.0.1/24",
		JailInterfaceName: "eth0",
		JailAddress:       "10.0.0.2/24",
		Gateway:           "10.0.0.1",
	}
	if err := vnet.Setup(&netCfg); err != nil {
		return fmt.Errorf("setup VNET: %w", err)
	}
	fmt.Printf("VNET configured: bridge=%s epair=%s<->%s\n", netCfg.Bridge, netCfg.EpairHost, netCfg.EpairJail)
	defer func() {
		if err := vnet.Teardown(&netCfg); err != nil {
			log.Printf("teardown: %v", err)
			return
		}
		fmt.Println("VNET torn down")
	}()

	out, err := jailexec.Output(jailName, "/sbin/ifconfig", "eth0")
	if err != nil {
		return fmt.Errorf("ifconfig inside jail: %w", err)
	}
	fmt.Printf("Inside the jail:\n%s", out)
	return nil
}
