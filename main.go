package main

import (
	"os"

	"github.com/blu3ph4ntom/port0/cmd"
)

var Version = "dev"

func main() {
	if os.Getenv("PORT0_DAEMON") == "1" {
		runDaemon()
		return
	}
	cmd.Execute(Version)
}
