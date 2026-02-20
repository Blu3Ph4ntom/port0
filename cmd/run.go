package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"

	"github.com/bluephantom/port0/internal/ipc"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	runName      string
	runRestart   string
	runTLS       bool
	runPortRange string
	runDetach    bool
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
		name := data["name"].(string)
		port := int(data["port"].(float64))
		url := data["url"].(string)

		green := color.New(color.FgGreen, color.Bold).SprintFunc()
		cyan := color.New(color.FgCyan).SprintFunc()
		yellow := color.New(color.FgYellow).SprintFunc()

		fmt.Println()
		fmt.Printf("  %s %s\n", green("✓"), "Server started")
		fmt.Println()
		fmt.Printf("  %s  %s\n", cyan("Name:"), name)
		fmt.Printf("  %s  %d\n", cyan("Port:"), port)
		fmt.Println()
		fmt.Printf("  %s\n", yellow("Access at:"))
		fmt.Printf("    • %s\n", url)
		fmt.Printf("    • http://%s.web\n", name)
		fmt.Printf("    • http://%s.local\n", name)
		fmt.Println()

		if runDetach {
			fmt.Printf("  Running in background. Use %s to view logs.\n", cyan(fmt.Sprintf("port0 logs %s", name)))
			fmt.Println()
			return nil
		}

		fmt.Printf("  %s\n", color.New(color.Faint).Sprint("Press Ctrl+C to stop"))
		fmt.Println()
		fmt.Println(color.New(color.Faint).Sprint("─────────────────────────────────────────────────────────"))
		fmt.Println()

		logsConn, err := ipc.Connect()
		if err != nil {
			return fmt.Errorf("error: cannot connect for logs: %w", err)
		}
		defer logsConn.Close()

		logsReq := ipc.LogsRequest{
			Name:   name,
			Follow: true,
		}
		if err := ipc.SendRequest(logsConn, "logs", logsReq); err != nil {
			return fmt.Errorf("error: %w", err)
		}

		scanner := bufio.NewScanner(logsConn)
		for scanner.Scan() {
			line := scanner.Text()
			var logLine map[string]string
			if err := json.Unmarshal([]byte(line), &logLine); err != nil {
				continue
			}
			if msg, ok := logLine["line"]; ok {
				fmt.Print(msg)
			}
		}

		return nil
	},
}

func init() {
	runCmd.Flags().StringVar(&runName, "name", "", "override project name (default: derived from cwd)")
	runCmd.Flags().StringVar(&runRestart, "restart", "no", "restart policy: no, always, on-failure")
	runCmd.Flags().BoolVar(&runTLS, "tls", false, "enable HTTPS for this project")
	runCmd.Flags().StringVar(&runPortRange, "port-range", "4000-4999", "port range (min-max)")
	runCmd.Flags().BoolVarP(&runDetach, "detach", "d", false, "run in background")
	rootCmd.AddCommand(runCmd)
}
