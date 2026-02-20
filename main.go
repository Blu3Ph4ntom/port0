package main

import (
	"os"

	"github.com/bluephantom/port0/cmd"
)

var Version = "dev"

func main() {
	if os.Getenv("PORT0_DAEMON") == "1" {
		runDaemon()
		return
	}
	cmd.Execute(Version)
}
