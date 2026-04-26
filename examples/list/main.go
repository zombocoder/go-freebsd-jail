package main

import (
	"fmt"
	"log"

	"github.com/zombocoder/go-freebsd-jail/jail"
)

func main() {
	jails, err := jail.List()
	if err != nil {
		log.Fatalf("Failed to list jails: %v", err)
	}

	if len(jails) == 0 {
		fmt.Println("No active jails")
		return
	}

	fmt.Printf("%-6s %-20s %-30s %s\n", "JID", "Name", "Hostname", "Path")
	fmt.Println("---    ----                 --------                       ----")
	for _, j := range jails {
		fmt.Printf("%-6d %-20s %-30s %s\n", j.JID, j.Name, j.Hostname, j.Path)
	}
}
