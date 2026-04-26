package main

import (
	"fmt"
	"log"

	"github.com/zombocoder/go-freebsd-jail/jail"
	"github.com/zombocoder/go-freebsd-jail/jailconf"
)

func main() {
	configs := map[string]jail.Config{
		"web": {
			Path:     "/jails/web",
			Hostname: "web.local",
			Persist:  true,
			VNET:     true,
			VNETInterface: []string{"epair0b"},
			MountDevFS: true,
			Allow: jail.AllowConfig{
				RawSockets: true,
			},
		},
		"db": {
			Path:     "/jails/db",
			Hostname: "db.local",
			Persist:  true,
			IP4:      []string{"10.0.0.3"},
			Allow: jail.AllowConfig{
				SysVIPC: true,
			},
		},
	}

	output, err := jailconf.RenderMany(configs)
	if err != nil {
		log.Fatalf("Failed to render jail.conf: %v", err)
	}

	fmt.Print(output)
}
