package main

import (
	"fmt"
	"os"

	"github.com/blu3ph4ntom/port0/internal/daemon"
)

func runDaemon() {
	if err := daemon.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: daemon: %v\n", err)
		os.Exit(1)
	}
}
