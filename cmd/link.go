package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/blu3ph4ntom/port0/internal/ipc"
	"github.com/spf13/cobra"
)

var linkPort int

var linkCmd = &cobra.Command{
	Use:   "link <name>",
	Short: "Link an existing server to a port0 name",
	Long:  "Creates a routing entry for a server that was not started via port0.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if linkPort == 0 {
			return fmt.Errorf("error: --port is required for link")
		}

		conn, err := ipc.Connect()
		if err != nil {
			return fmt.Errorf("port0 daemon not running. start it with: port0 daemon start")
		}
		defer conn.Close()

		cwd, _ := os.Getwd()
		req := ipc.LinkRequest{
			Name: args[0],
			Port: linkPort,
			Cwd:  cwd,
		}

		if err := ipc.SendRequest(conn, "link", req); err != nil {
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
		fmt.Printf("linked %s -> %s (port %v)\n", data["name"], data["url"], data["port"])

		return nil
	},
}

func init() {
	linkCmd.Flags().IntVar(&linkPort, "port", 0, "port the existing server is running on")
	linkCmd.MarkFlagRequired("port")
	rootCmd.AddCommand(linkCmd)
}
