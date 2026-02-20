package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/bluephantom/port0/internal/ipc"
	"github.com/spf13/cobra"
)

var killRemove bool

var killCmd = &cobra.Command{
	Use:   "kill <name>",
	Short: "Stop a running project",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		conn, err := ipc.Connect()
		if err != nil {
			return fmt.Errorf("port0 daemon not running. start it with: port0 daemon start")
		}
		defer conn.Close()

		req := ipc.KillRequest{
			Name:   args[0],
			Remove: killRemove,
		}

		if err := ipc.SendRequest(conn, "kill", req); err != nil {
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

		if data["removed"] == true {
			fmt.Printf("removed %s\n", data["name"])
		} else {
			fmt.Printf("killed %s (pid %v)\n", data["name"], data["pid"])
		}

		return nil
	},
}

func init() {
	killCmd.Flags().BoolVar(&killRemove, "rm", false, "also remove from state")
	rootCmd.AddCommand(killCmd)
}
