//go:build windows

package engine

import (
	"os"
	"os/exec"
)

func ExecShim(binary string, args []string) error {
	cmd := exec.Command(binary, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
