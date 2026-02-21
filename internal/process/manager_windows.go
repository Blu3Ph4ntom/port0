//go:build windows

package process

import (
	"os/exec"
	"syscall"
)

const (
	CREATE_NO_WINDOW         = 0x08000000
	CREATE_NEW_PROCESS_GROUP = 0x00000200
	DETACHED_PROCESS         = 0x00000008
)

func setChildProcessAttrs(cmd *exec.Cmd) {
	if cmd.SysProcAttr == nil {
		cmd.SysProcAttr = &syscall.SysProcAttr{}
	}
	// Combine flags: DETACHED_PROCESS creates a process that is detached from the console
	// CREATE_NO_WINDOW ensures no window is created
	// CREATE_NEW_PROCESS_GROUP puts the process in its own group
	cmd.SysProcAttr.CreationFlags = CREATE_NO_WINDOW | CREATE_NEW_PROCESS_GROUP | DETACHED_PROCESS
	cmd.SysProcAttr.HideWindow = true
}
