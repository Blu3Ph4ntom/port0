package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/blu3ph4ntom/port0/internal/ipc"
	"github.com/spf13/cobra"
)

var logsFollow bool

var logsCmd = &cobra.Command{
	Use:   "logs <name>",
	Short: "View logs for a project",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		conn, err := ipc.Connect()
		if err != nil {
			return fmt.Errorf("port0 daemon not running. start it with: port0 daemon start")
		}
		defer conn.Close()

		req := ipc.LogsRequest{
			Name:   args[0],
			Follow: logsFollow,
		}

		if err := ipc.SendRequest(conn, "logs", req); err != nil {
			return fmt.Errorf("error: %w", err)
		}

		if logsFollow {
			err := ipc.StreamLines(conn, func(line, ts string) bool {
				fmt.Print(line)
				return true
			})
			if err != nil {
				return fmt.Errorf("error: %w", err)
			}
			return nil
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
		if lines, ok := data["lines"].([]interface{}); ok {
			for _, line := range lines {
				fmt.Println(line)
			}
		}

		return nil
	},
}

func init() {
	logsCmd.Flags().BoolVarP(&logsFollow, "follow", "f", false, "stream live logs")
	rootCmd.AddCommand(logsCmd)
}
