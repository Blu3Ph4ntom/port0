package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"

	"github.com/blu3ph4ntom/port0/internal/ipc"
	"github.com/blu3ph4ntom/port0/internal/util"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	runName      string
	runDomain    string
	runRestart   string
	runTLS       bool
	runPortRange string
	runDetach    bool
)

var runCmd = &cobra.Command{
	Use:   "run [flags] <cmd...>",
	Short: "Run a dev server with port0",
	Long:  "Wraps a command, injects PORT, and proxies it under a human-readable hostname.",
	Args:  cobra.MinimumNArgs(1),
	RunE:  runServer,
}

func runServer(cmd *cobra.Command, args []string) error {
	if err := ensureDaemon(); err != nil {
		return fmt.Errorf("error: %w", err)
	}

	cwd, _ := os.Getwd()

	// Parse name.domain format (e.g., "api.myapp" -> name="api", domain="myapp")
	name, domain := parseNameAndDomain(runName, runDomain)

	// If not detached, run in foreground mode
	if !runDetach {
		return runForeground(args, cwd, name, domain)
	}

	// Detached mode: spawn via daemon
	req := ipc.SpawnRequest{
		Name:      name,
		Cmd:       args,
		Cwd:       cwd,
		Restart:   runRestart,
		TLS:       runTLS,
		PortRange: runPortRange,
		Domain:    domain,
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
	if err := json.Unmarshal(resp.Data, &data); err != nil {
		return fmt.Errorf("error: failed to parse response: %w", err)
	}
	respName := data["name"].(string)
	port := int(data["port"].(float64))
	url := data["url"].(string)

	// Build base host for additional URLs
	baseHost := respName
	if respDomain, ok := data["domain"].(string); ok && respDomain != "" {
		baseHost = respName + "." + respDomain
	}

	cyan := color.New(color.FgCyan).SprintFunc()
	dim := color.New(color.FgHiBlack).SprintFunc()
	fmt.Printf("port0: %s → %s (port %d)\n", respName, cyan(url), port)
	fmt.Printf("  %s %s, %s\n", dim("also:"), cyan(fmt.Sprintf("http://%s.web", baseHost)), cyan(fmt.Sprintf("http://%s.local", baseHost)))
	fmt.Printf("port0: use %s to view logs\n", cyan(fmt.Sprintf("port0 logs %s", respName)))

	return nil
}

func runForeground(args []string, cwd string, name, domain string) error {
	// Get a port allocation from daemon
	conn, err := ipc.Connect()
	if err != nil {
		return fmt.Errorf("error: cannot connect to daemon: %w", err)
	}

	if name == "" {
		name = util.FromCwd(cwd)
	}

	req := ipc.RegisterRequest{
		Name:      name,
		Cmd:       args,
		Cwd:       cwd,
		PortRange: runPortRange,
		Domain:    domain,
	}

	if err := ipc.SendRequest(conn, "register", req); err != nil {
		conn.Close()
		return fmt.Errorf("error: %w", err)
	}

	resp, err := ipc.ReadResponse(conn)
	conn.Close()
	if err != nil {
		return fmt.Errorf("error: %w", err)
	}

	if !resp.OK {
		return fmt.Errorf("error: %s", resp.Error)
	}

	var data map[string]interface{}
	if err := json.Unmarshal(resp.Data, &data); err != nil {
		return fmt.Errorf("error: failed to parse response: %w", err)
	}
	port := int(data["port"].(float64))
	url := data["url"].(string)

	// Build base host for additional URLs
	baseHost := name
	if domain != "" {
		baseHost = name + "." + domain
	}

	// Print minimal info
	cyan := color.New(color.FgCyan).SprintFunc()
	dim := color.New(color.FgHiBlack).SprintFunc()
	fmt.Printf("port0: %s → %s (port %d)\n", name, cyan(url), port)
	fmt.Printf("  %s %s, %s\n", dim("also:"), cyan(fmt.Sprintf("http://%s.web", baseHost)), cyan(fmt.Sprintf("http://%s.local", baseHost)))

	// Setup signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Start the process with PORT env
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Dir = cwd
	cmd.Env = append(os.Environ(), fmt.Sprintf("PORT=%d", port))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Start(); err != nil {
		// Unregister on failure
		unregister(name)
		return fmt.Errorf("error: failed to start: %w", err)
	}

	// Wait for signal or process exit
	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	select {
	case <-sigChan:
		// User pressed Ctrl+C
		cmd.Process.Signal(os.Interrupt)
		<-done
	case err := <-done:
		// Process exited on its own
		if err != nil {
			fmt.Fprintf(os.Stderr, "port0: process exited with error: %v\n", err)
		}
	}

	// Unregister from daemon
	unregister(name)
	return nil
}

func unregister(name string) {
	conn, err := ipc.Connect()
	if err != nil {
		return
	}
	defer conn.Close()

	req := map[string]string{"name": name}
	ipc.SendRequest(conn, "unregister", req)
}

// parseNameAndDomain parses name in formats:
//   - "api.myapp" -> ("api", "myapp")  // dot notation
//   - "api" -> ("api", "")             // simple name
//
// If explicitDomain is provided, it takes precedence over dot notation.
func parseNameAndDomain(name, explicitDomain string) (string, string) {
	if explicitDomain != "" {
		return name, explicitDomain
	}
	if idx := strings.Index(name, "."); idx != -1 {
		return name[:idx], name[idx+1:]
	}
	return name, ""
}

func init() {
	runCmd.Flags().StringVarP(&runName, "name", "n", "", "custom project name (default: from folder)")
	runCmd.Flags().StringVar(&runDomain, "domain", "", "parent domain for subdomain routing (e.g., 'myapp' for <name>.myapp.localhost)")
	runCmd.Flags().StringVar(&runRestart, "restart", "no", "restart policy: no, always, on-failure")
	runCmd.Flags().BoolVar(&runTLS, "tls", false, "enable HTTPS for this project")
	runCmd.Flags().StringVar(&runPortRange, "port-range", "4000-4999", "port range (min-max)")
	runCmd.Flags().BoolVarP(&runDetach, "detach", "d", false, "run in background")
	rootCmd.AddCommand(runCmd)
}
