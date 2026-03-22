package engine

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// List returns all installed Godot versions, sorted alphabetically.
// The default version (from config) is marked with IsDefault=true.
func (s *Service) List() ([]InstalledVersion, error) {
	names, err := listDirectories(s.VersionsDir())
	if err != nil {
		return nil, err
	}
	var versions []InstalledVersion
	for _, name := range names {
		versions = append(versions, InstalledVersion{
			Version:   name,
			IsDefault: name == s.Config.DefaultVersion,
		})
	}
	return versions, nil
}

// ListVersionStrings returns the version strings of all installed versions.
func (s *Service) ListVersionStrings() ([]string, error) {
	versions, err := s.List()
	if err != nil {
		return nil, err
	}
	out := make([]string, len(versions))
	for i, v := range versions {
		out[i] = v.Version
	}
	return out, nil
}

// IsInstalled checks whether the given version is installed.
func (s *Service) IsInstalled(version string) bool {
	_, err := os.Stat(filepath.Join(s.VersionsDir(), version))
	return err == nil
}

// BinaryPath returns the absolute path to the Godot binary for the given version.
// It first checks for a canonical name ("godot" / "godot.exe"), then scans for
// the Godot_v* pattern. On Windows, _console variants are skipped.
func (s *Service) BinaryPath(version string) (string, error) {
	versionDir := filepath.Join(s.VersionsDir(), version)

	// Try canonical name first
	canonical := "godot"
	if s.Platform.OS == "windows" {
		canonical = "godot.exe"
	}
	p := filepath.Join(versionDir, canonical)
	if _, err := os.Stat(p); err == nil {
		return p, nil
	}

	// Scan for actual Godot binary (e.g. Godot_v4.3-stable_linux.x86_64)
	entries, err := os.ReadDir(versionDir)
	if err != nil {
		return "", &ActionableError{
			Err:        fmt.Errorf("version %s is not installed", version),
			Suggestion: fmt.Sprintf("gdt install %s", version),
		}
	}

	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if strings.HasPrefix(name, "Godot_v") {
			if s.Platform.OS == "windows" && strings.Contains(name, "_console") {
				continue
			}
			return filepath.Join(versionDir, name), nil
		}
	}

	return "", &ActionableError{
		Err:        fmt.Errorf("engine binary not found in %s", versionDir),
		Suggestion: fmt.Sprintf("gdt install %s --force", version),
	}
}
