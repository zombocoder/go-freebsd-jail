package main

import (
	"fmt"
	"log"

	"github.com/zombocoder/go-freebsd-jail/jail"
)

func main() {
	jid, err := jail.Create(jail.Config{
		Name:     "example-basic",
		Path:     "/jails/example-basic",
		Hostname: "example-basic.local",
		Persist:  true,
	})
	if err != nil {
		log.Fatalf("Failed to create jail: %v", err)
	}
	fmt.Printf("Created jail with JID %d\n", jid)

	// List all jails
	jails, err := jail.List()
	if err != nil {
		log.Fatalf("Failed to list jails: %v", err)
	}
	for _, j := range jails {
		fmt.Printf("  JID=%d Name=%s Hostname=%s Path=%s\n", j.JID, j.Name, j.Hostname, j.Path)
	}

	// Clean up
	if err := jail.Remove("example-basic"); err != nil {
		log.Fatalf("Failed to remove jail: %v", err)
	}
	fmt.Println("Jail removed")
}
