package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/bluephantom/port0/internal/ipc"
	"github.com/spf13/cobra"
)

var lsCmd = &cobra.Command{
	Use:   "ls",
	Short: "List all port0 projects",
	RunE: func(cmd *cobra.Command, args []string) error {
		conn, err := ipc.Connect()
		if err != nil {
			return fmt.Errorf("port0 daemon not running. start it with: port0 daemon start")
		}
		defer conn.Close()

		if err := ipc.SendRequest(conn, "ls", nil); err != nil {
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

		var projects []map[string]interface{}
		json.Unmarshal(resp.Data, &projects)

		if len(projects) == 0 {
			fmt.Println("no projects")
			return nil
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
		fmt.Fprintf(w, "NAME\tPORT\tURL\tPID\tSTATUS\tSTARTED\n")

		for _, p := range projects {
			name := p["name"].(string)
			port := int(p["port"].(float64))
			url := p["url"].(string)
			pid := int(p["pid"].(float64))
			status := p["status"].(string)

			started := ""
			if ts, ok := p["started_at"].(string); ok && ts != "" {
				t, err := time.Parse(time.RFC3339Nano, ts)
				if err == nil && !t.IsZero() {
					started = humanDuration(time.Since(t))
				}
			}

			fmt.Fprintf(w, "%s\t%d\t%s\t%d\t%s\t%s\n", name, port, url, pid, status, started)
		}
		w.Flush()

		return nil
	},
}

func init() {
	rootCmd.AddCommand(lsCmd)
}

func humanDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%ds ago", int(d.Seconds()))
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm ago", int(d.Minutes()))
	}
	if d < 24*time.Hour {
		return fmt.Sprintf("%dh ago", int(d.Hours()))
	}
	return fmt.Sprintf("%dd ago", int(d.Hours()/24))
}
