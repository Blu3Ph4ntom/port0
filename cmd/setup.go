package cmd

import (
	"fmt"

	"github.com/bluephantom/port0/internal/setup"
	"github.com/spf13/cobra"
)

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Configure DNS and system services for port0",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := setup.Setup(); err != nil {
			return fmt.Errorf("error: %w", err)
		}
		return nil
	},
}

var teardownCmd = &cobra.Command{
	Use:   "teardown",
	Short: "Remove port0 system configuration",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := setup.Teardown(); err != nil {
			return fmt.Errorf("error: %w", err)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(setupCmd)
	rootCmd.AddCommand(teardownCmd)
}
