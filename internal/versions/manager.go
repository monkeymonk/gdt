package versions

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
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
	p := filepath.Join(versionsDir, BinaryPath(version, goos))
	if _, err := os.Stat(p); err != nil {
		return "", errors.New(fmt.Sprintf("engine binary not found: %s\n\n  gdt install %s", p, version))
	}
	return p, nil
}
