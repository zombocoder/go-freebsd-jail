package main

import (
	"fmt"
	"log"

	"github.com/zombocoder/go-freebsd-jail/jail"
	"github.com/zombocoder/go-freebsd-jail/vnet"
)

func main() {
	// Create a VNET jail
	jid, err := jail.Create(jail.Config{
		Name:     "example-vnet",
		Path:     "/jails/example-vnet",
		Hostname: "example-vnet.local",
		Persist:  true,
		VNET:     true,
	})
	if err != nil {
		log.Fatalf("Failed to create jail: %v", err)
	}
	fmt.Printf("Created VNET jail with JID %d\n", jid)

	// Set up networking
	netCfg := vnet.NetworkConfig{
		JailName:          "example-vnet",
		Bridge:            "bridge0",
		CreateEpair:       true,
		JailInterfaceName: "eth0",
		JailAddress:       "10.0.0.2/24",
		Gateway:           "10.0.0.1",
	}

	if err := vnet.Setup(netCfg); err != nil {
		log.Fatalf("Failed to setup VNET: %v", err)
	}
	fmt.Println("VNET networking configured")

	// Teardown and remove
	if err := vnet.Teardown(netCfg); err != nil {
		log.Printf("Warning: teardown error: %v", err)
	}
	if err := jail.Remove("example-vnet"); err != nil {
		log.Fatalf("Failed to remove jail: %v", err)
	}
	fmt.Println("Jail removed")
}
