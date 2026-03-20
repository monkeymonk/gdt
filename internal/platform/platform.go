package platform

import (
	"os"
	"path/filepath"
	"runtime"
)

type Info struct {
	OS   string
	Arch string
}

func Detect() Info {
	return Info{
		OS:   runtime.GOOS,
		Arch: runtime.GOARCH,
	}
}

func (p Info) DefaultHome() string {
	switch p.OS {
	case "darwin":
		home, _ := os.UserHomeDir()
		return filepath.Join(home, "Library", "Application Support", "gdt")
	case "windows":
		return filepath.Join(os.Getenv("LOCALAPPDATA"), "gdt")
	default:
		home, _ := os.UserHomeDir()
		return filepath.Join(home, ".gdt")
	}
}
