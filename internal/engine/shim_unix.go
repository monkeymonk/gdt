//go:build !windows

package engine

import (
	"os"
	"syscall"
)

func ExecShim(binary string, args []string) error {
	argv := append([]string{binary}, args...)
	return syscall.Exec(binary, argv, os.Environ())
}
