package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/bluephantom/port0/internal/ipc"
	"github.com/spf13/cobra"
)

var (
	runName      string
	runRestart   string
	runTLS       bool
	runPortRange string
)

var runCmd = &cobra.Command{
	Use:   "run [flags] -- <cmd...>",
	Short: "Run a dev server with port0",
	Long:  "Wraps a command, injects PORT, and proxies it under a human-readable hostname.",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := ensureDaemon(); err != nil {
			return fmt.Errorf("error: %w", err)
		}

		cwd, _ := os.Getwd()
		req := ipc.SpawnRequest{
			Name:      runName,
			Cmd:       args,
			Cwd:       cwd,
			Restart:   runRestart,
			TLS:       runTLS,
			PortRange: runPortRange,
		}

		conn, err := ipc.Connect()
		if err != nil {
			return fmt.Errorf("error: cannot connect to daemon: %w", err)
		}
		defer conn.Close()

		if err := ipc.SendRequest(conn, "spawn", req); err != nil {
			return fmt.Errorf("error: %w", err)
		}

		resp, err := ipc.ReadResponse(conn)
		if err != nil {
			return fmt.Errorf("error: %w", err)
		}

		if !resp.OK {
			return fmt.Errorf("error: %s", resp.Error)
		}

		if jsonOutput {
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			enc.Encode(resp.Data)
			return nil
		}

		var data map[string]interface{}
		json.Unmarshal(resp.Data, &data)
		fmt.Printf("name:  %s\n", data["name"])
		fmt.Printf("port:  %v\n", data["port"])
		fmt.Printf("url:   %s\n", data["url"])
		fmt.Printf("cmd:   %v\n", data["cmd"])
		if data["tls"] == true {
			fmt.Printf("note:  self-signed cert at ~/.port0/certs/%s.pem\n", data["name"])
		}

		return nil
	},
}

func init() {
	runCmd.Flags().StringVar(&runName, "name", "", "override project name (default: derived from cwd)")
	runCmd.Flags().StringVar(&runRestart, "restart", "no", "restart policy: no, always, on-failure")
	runCmd.Flags().BoolVar(&runTLS, "tls", false, "enable HTTPS for this project")
	runCmd.Flags().StringVar(&runPortRange, "port-range", "4000-4999", "port range (min-max)")
	rootCmd.AddCommand(runCmd)
}
