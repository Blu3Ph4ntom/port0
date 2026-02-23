//go:build !windows

package process

import (
	"os"
	"syscall"
)

// IsPIDRunning reports whether the given PID appears to be running on Unix-like systems.
//
// Notes:
//   - pid <= 0 is always treated as not running.
//   - On Unix, os.FindProcess does not validate existence; the Signal(0) probe does.
func IsPIDRunning(pid int) bool {
	if pid <= 0 {
		return false
	}

	proc, err := os.FindProcess(pid)
	if err != nil {
		return false
	}

	// Signal 0 does not send a signal, but performs error checking.
	err = proc.Signal(syscall.Signal(0))
	return err == nil
}
