package main

import (
	"fmt"
	"log"
	"os"

	"github.com/zombocoder/go-freebsd-jail/jail"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <jail-name>\n", os.Args[0])
		os.Exit(1)
	}

	name := os.Args[1]
	if err := jail.Remove(name); err != nil {
		log.Fatalf("Failed to remove jail %q: %v", name, err)
	}
	fmt.Printf("Jail %q removed\n", name)
}
