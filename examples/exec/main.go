package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	jailexec "github.com/zombocoder/go-freebsd-jail/exec"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Fprintf(os.Stderr, "Usage: %s <jail-name> <command> [args...]\n", os.Args[0])
		os.Exit(1)
	}

	jailName := os.Args[1]
	cmdPath := os.Args[2]
	cmdArgs := os.Args[3:]

	out, err := jailexec.Output(jailName, cmdPath, cmdArgs...)
	if err != nil {
		log.Fatalf("Failed to execute in jail %q: %v", jailName, err)
	}
	fmt.Print(strings.TrimRight(string(out), "\n"))
}
