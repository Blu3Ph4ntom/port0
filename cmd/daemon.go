package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/blu3ph4ntom/port0/internal/ipc"
	"github.com/spf13/cobra"
)

var daemonCmd = &cobra.Command{
	Use:   "daemon",
	Short: "Manage the port0 daemon",
}

var daemonStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the port0 daemon in the background",
	RunE: func(cmd *cobra.Command, args []string) error {
		if isDaemonRunning() {
			pid := readDaemonPid()
			fmt.Printf("daemon already running (pid %d)\n", pid)
			return nil
		}

		if err := startDaemonProcess(); err != nil {
			return err
		}
		return nil
	},
}

var daemonStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop the port0 daemon",
	RunE: func(cmd *cobra.Command, args []string) error {
		pid := readDaemonPid()
		if pid == 0 {
			fmt.Println("daemon is not running")
			return nil
		}

		proc, err := os.FindProcess(pid)
		if err != nil {
			fmt.Println("daemon is not running")
			os.Remove(ipc.PidPath())
			return nil
		}

		if err := proc.Signal(syscall.SIGTERM); err != nil {
			proc.Kill()
		}

		os.Remove(ipc.PidPath())
		os.Remove(ipc.SocketPath())
		fmt.Printf("daemon stopped (pid %d)\n", pid)
		return nil
	},
}

var daemonStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check daemon status",
	RunE: func(cmd *cobra.Command, args []string) error {
		if !isDaemonRunning() {
			fmt.Println("not running")
			return nil
		}

		conn, err := ipc.Connect()
		if err != nil {
			pid := readDaemonPid()
			if pid > 0 {
				fmt.Printf("running (pid %d) but socket unreachable\n", pid)
			} else {
				fmt.Println("not running")
			}
			return nil
		}
		defer conn.Close()

		ipc.SendRequest(conn, "status", nil)
		resp, err := ipc.ReadResponse(conn)
		if err != nil {
			fmt.Println("running but unresponsive")
			return nil
		}

		if jsonOutput {
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			enc.Encode(resp.Data)
			return nil
		}

		var data map[string]interface{}
		json.Unmarshal(resp.Data, &data)
		pid := int(data["pid"].(float64))
		projects := int(data["projects"].(float64))
		running := int(data["running"].(float64))
		fmt.Printf("running (pid %d)\n", pid)
		fmt.Printf("  projects: %d (%d running)\n", projects, running)
		return nil
	},
}

func init() {
	daemonCmd.AddCommand(daemonStartCmd)
	daemonCmd.AddCommand(daemonStopCmd)
	daemonCmd.AddCommand(daemonStatusCmd)
	rootCmd.AddCommand(daemonCmd)
}

func isDaemonRunning() bool {
	pid := readDaemonPid()
	if pid == 0 {
		return false
	}
	return isProcessRunning(pid)
}

func readDaemonPid() int {
	data, err := os.ReadFile(ipc.PidPath())
	if err != nil {
		return 0
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil {
		return 0
	}
	return pid
}

func startDaemonProcess() error {
	bin, err := os.Executable()
	if err != nil {
		return fmt.Errorf("cannot find executable: %w", err)
	}

	child := exec.Command(bin, "daemon", "start")
	child.Env = append(os.Environ(), "PORT0_DAEMON=1")
	child.Stdin = nil
	child.Stdout = nil
	child.Stderr = nil
	child.SysProcAttr = daemonSysProcAttr()

	if err := child.Start(); err != nil {
		return fmt.Errorf("failed to start daemon: %w", err)
	}

	fmt.Printf("daemon started (pid %d)\n", child.Process.Pid)
	child.Process.Release()
	return nil
}

func ensureDaemon() error {
	if isDaemonRunning() {
		return nil
	}

	if err := startDaemonProcess(); err != nil {
		return err
	}

	for i := 0; i < 20; i++ {
		time.Sleep(100 * time.Millisecond)
		if _, err := os.Stat(ipc.SocketPath()); err == nil {
			return nil
		}
	}
	return fmt.Errorf("daemon started but socket not ready")
}
