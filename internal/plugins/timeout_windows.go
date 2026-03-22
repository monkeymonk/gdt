//go:build windows

package plugins

import "os/exec"

func setProcAttr(cmd *exec.Cmd) {
	// Windows does not support process groups via Setpgid.
	// Context cancellation will kill the process directly.
}
