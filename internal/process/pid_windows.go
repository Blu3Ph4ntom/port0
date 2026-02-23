//go:build windows

package process

import "syscall"

// IsPIDRunning returns true if the given PID appears to refer to a live process.
//
// Notes:
//   - This is best-effort on Windows. For port0's use (its own spawned children),
//     OpenProcess is sufficient and avoids heavy enumeration APIs.
//   - If pid <= 0, it returns false.
func IsPIDRunning(pid int) bool {
	if pid <= 0 {
		return false
	}

	// Try to open the process with minimal query rights.
	// PROCESS_QUERY_INFORMATION is widely available; combine with SYNCHRONIZE
	// to improve compatibility across Windows versions.
	h, err := syscall.OpenProcess(syscall.PROCESS_QUERY_INFORMATION|syscall.SYNCHRONIZE, false, uint32(pid))
	if err != nil {
		return false
	}
	_ = syscall.CloseHandle(h)
	return true
}
