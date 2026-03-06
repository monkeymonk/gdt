//go:build !windows

package shim

import (
	"os"
	"syscall"
)

func Exec(binary string, args []string) error {
	argv := append([]string{binary}, args...)
	return syscall.Exec(binary, argv, os.Environ())
}
