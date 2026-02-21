//go:build !windows

package process

import "os/exec"

func setChildProcessAttrs(cmd *exec.Cmd) {
	// No special attributes needed on Unix
}
