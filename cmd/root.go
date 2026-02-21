package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var (
	jsonOutput bool
	verbose    bool
	version    string
)

func port0Dir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		home = "."
	}
	return filepath.Join(home, ".port0")
}

var rootCmd = &cobra.Command{
	Use:           "port0",
	Short:         "No ports. Just names.",
	Long:          "port0 wraps any dev server, auto-assigns a free port, and reverse-proxies it under a human-readable hostname.",
	SilenceUsage:  true,
	SilenceErrors: true,
}

func init() {
	rootCmd.PersistentFlags().BoolVar(&jsonOutput, "json", false, "machine-readable JSON output")
	rootCmd.PersistentFlags().BoolVar(&verbose, "verbose", false, "show debug logs")
	rootCmd.PersistentFlags().StringVarP(&runName, "name", "n", "", "custom project name")
	rootCmd.PersistentFlags().BoolVarP(&runDetach, "detach", "d", false, "run in background")

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

	// Custom command handling
	if len(os.Args) > 1 {
		// Parse flags manually to separate them from the command
		flags := pflag.NewFlagSet("port0", pflag.ContinueOnError)
		flags.ParseErrorsWhitelist.UnknownFlags = true
		flags.BoolP("name", "n", false, "")
		flags.BoolP("detach", "d", false, "")
		flags.Bool("json", false, "")
		flags.Bool("verbose", false, "")

		// Find where the actual command starts (after all flags)
		cmdStartIdx := -1
		for i := 1; i < len(os.Args); i++ {
			arg := os.Args[i]
			if !strings.HasPrefix(arg, "-") {
				// Check if previous arg was a flag that takes a value
				if i > 1 && (os.Args[i-1] == "-n" || os.Args[i-1] == "--name") {
					continue
				}
				cmdStartIdx = i
				break
			}
		}

		if cmdStartIdx > 0 {
			cmdName := os.Args[cmdStartIdx]

			// Check if it's a known subcommand
			isKnownCmd := false
			for _, c := range rootCmd.Commands() {
				if c.Name() == cmdName || c.HasAlias(cmdName) {
					isKnownCmd = true
					break
				}
			}

			// If not a known command, run as server
			if !isKnownCmd && cmdName != "help" && cmdName != "version" {
				// Parse our flags first
				rootCmd.PersistentFlags().Parse(os.Args[1:cmdStartIdx])

				// Call runServer with remaining args
				if err := runServer(nil, os.Args[cmdStartIdx:]); err != nil {
					fmt.Fprintf(os.Stderr, "error: %v\n", err)
					os.Exit(1)
				}
				return
			}
		}
	}

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
