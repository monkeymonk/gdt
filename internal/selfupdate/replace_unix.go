//go:build unix

package selfupdate

import "os"

func replaceBinary(dst, src string) error {
	return os.Rename(src, dst)
}
