package config

import (
	"os"
	"path/filepath"
	"runtime"
)

// DefaultHome returns the default gdt home directory for the current OS.
func DefaultHome() string {
	switch runtime.GOOS {
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

// ResolveHome returns GDT_HOME env var if set, otherwise DefaultHome().
func ResolveHome() string {
	if env := os.Getenv("GDT_HOME"); env != "" {
		return env
	}
	return DefaultHome()
}
