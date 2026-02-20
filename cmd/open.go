package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"runtime"

	"github.com/blu3ph4ntom/port0/internal/ipc"
	"github.com/spf13/cobra"
)

var openCmd = &cobra.Command{
	Use:   "open <name>",
	Short: "Open a project in the browser",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		conn, err := ipc.Connect()
		if err != nil {
			return fmt.Errorf("port0 daemon not running. start it with: port0 daemon start")
		}
		defer conn.Close()

		req := ipc.OpenRequest{Name: args[0]}
		if err := ipc.SendRequest(conn, "open", req); err != nil {
			return fmt.Errorf("error: %w", err)
		}

		resp, err := ipc.ReadResponse(conn)
		if err != nil {
			return fmt.Errorf("error: %w", err)
		}

		if !resp.OK {
			return fmt.Errorf("error: %s", resp.Error)
		}

		var data map[string]interface{}
		json.Unmarshal(resp.Data, &data)
		url := data["url"].(string)

		if jsonOutput {
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			enc.Encode(resp.Data)
			return nil
		}

		fmt.Printf("opening %s\n", url)
		return openBrowser(url)
	},
}

func init() {
	rootCmd.AddCommand(openCmd)
}

func openBrowser(url string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", url)
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
	return cmd.Start()
}
