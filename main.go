package main

import (
	"github.com/certfix/certfix-cli/cmd/certfix"
)

var (
	// Version is set during build time via ldflags
	Version = "dev"
)

func main() {
	// Set version in certfix package
	if Version != "dev" {
		certfix.Version = Version
	}
	certfix.Execute()
}
