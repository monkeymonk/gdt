package versions

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func List(versionsDir string) ([]string, error) {
	entries, err := os.ReadDir(versionsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var versions []string
	for _, e := range entries {
		if e.IsDir() {
			versions = append(versions, e.Name())
		}
	}
	sort.Strings(versions)
	return versions, nil
}

func IsInstalled(versionsDir string, version string) bool {
	_, err := os.Stat(filepath.Join(versionsDir, version))
	return err == nil
}

func Remove(versionsDir string, version string) error {
	dir := filepath.Join(versionsDir, version)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return fmt.Errorf("version %s is not installed", version)
	}
	return os.RemoveAll(dir)
}

func BinaryPath(version string, goos string) string {
	name := "godot"
	if goos == "windows" {
		name = "godot.exe"
	}
	return filepath.Join(version, name)
}

func AbsoluteBinaryPath(versionsDir string, version string, goos string) (string, error) {
	// Try canonical name first
	p := filepath.Join(versionsDir, BinaryPath(version, goos))
	if _, err := os.Stat(p); err == nil {
		return p, nil
	}

	// Scan for actual Godot binary (e.g. Godot_v4.3-stable_linux.x86_64)
	versionDir := filepath.Join(versionsDir, version)
	entries, err := os.ReadDir(versionDir)
	if err != nil {
		return "", fmt.Errorf("version %s is not installed\n\n  gdt install %s", version, version)
	}

	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if strings.HasPrefix(name, "Godot_v") {
			return filepath.Join(versionDir, name), nil
		}
	}

	return "", fmt.Errorf("engine binary not found in %s\n\n  gdt install %s --force", versionDir, version)
}
