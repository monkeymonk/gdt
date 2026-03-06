package shim

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func IsShimInvocation(argv0 string) bool {
	base := filepath.Base(argv0)
	base = strings.TrimSuffix(base, ".exe")
	return base == "godot"
}

func CreateShimLink(gdtBinaryPath string, shimsDir string) error {
	if err := os.MkdirAll(shimsDir, 0755); err != nil {
		return err
	}
	link := filepath.Join(shimsDir, "godot")
	os.Remove(link)
	if err := os.Symlink(gdtBinaryPath, link); err != nil {
		return fmt.Errorf("failed to create shim symlink: %w", err)
	}
	return nil
}
