package engine

import (
	"os"
	"path/filepath"
	"strings"
)

func IsShimInvocation(argv0 string) bool {
	base := filepath.Base(argv0)
	base = strings.TrimSuffix(base, ".exe")
	return base == "godot"
}

func (s *Service) CreateShimLink() error {
	if err := os.MkdirAll(s.shimsDir(), 0o755); err != nil {
		return err
	}
	gdtBin, err := os.Executable()
	if err != nil {
		return err
	}
	gdtBin, err = filepath.EvalSymlinks(gdtBin)
	if err != nil {
		return err
	}
	link := filepath.Join(s.shimsDir(), "godot")
	_ = os.Remove(link)
	return os.Symlink(gdtBin, link)
}
