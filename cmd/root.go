package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var (
	jsonOutput bool
	verbose    bool
	version    string
)

func port0Dir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".port0")
}

var rootCmd = &cobra.Command{
	Use:   "port0",
	Short: "No ports. Just names.",
	Long:  "port0 wraps any dev server, auto-assigns a free port, and reverse-proxies it under a human-readable hostname.",
	SilenceUsage:  true,
	SilenceErrors: true,
}

func init() {
	rootCmd.PersistentFlags().BoolVar(&jsonOutput, "json", false, "machine-readable JSON output")
	rootCmd.PersistentFlags().BoolVar(&verbose, "verbose", false, "show debug logs")

	cobra.OnInitialize(initDirs)
}

func initDirs() {
	base := port0Dir()
	dirs := []string{
		base,
		filepath.Join(base, "logs"),
		filepath.Join(base, "certs"),
	}
	for _, d := range dirs {
		os.MkdirAll(d, 0755)
	}
}

func Execute(v string) {
	version = v
	rootCmd.Version = v
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
