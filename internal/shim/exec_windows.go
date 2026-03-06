//go:build windows

package shim

import (
	"os"
	"os/exec"
)

func Exec(binary string, args []string) error {
	cmd := exec.Command(binary, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
