//go:build windows

package selfupdate

import "os"

func replaceBinary(dst, src string) error {
	oldPath := dst + ".old"
	if _, err := os.Stat(dst); err == nil {
		os.Rename(dst, oldPath)
	}

	err := os.Rename(src, dst)
	os.Remove(oldPath)
	return err
}
