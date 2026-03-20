package engine

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Resolve determines which Godot version to use by checking sources in order:
// 1. .godot-version file (walks parent dirs from startDir)
// 2. GDT_GODOT_VERSION env var
// 3. Config default version
// 4. Latest installed version
func (s *Service) Resolve(startDir string) (ResolvedVersion, error) {
	// 1. .godot-version file
	if v, err := resolveFromFile(startDir); err == nil {
		bin, binErr := s.BinaryPath(v)
		return ResolvedVersion{Version: v, BinaryPath: bin, Source: "file"}, binErr
	}

	// 2. Environment variable
	if v := os.Getenv("GDT_GODOT_VERSION"); v != "" {
		bin, binErr := s.BinaryPath(v)
		return ResolvedVersion{Version: v, BinaryPath: bin, Source: "env"}, binErr
	}

	// 3. Config default
	if s.Config.DefaultVersion != "" {
		v := s.Config.DefaultVersion
		bin, binErr := s.BinaryPath(v)
		return ResolvedVersion{Version: v, BinaryPath: bin, Source: "config"}, binErr
	}

	// 4. Latest installed
	versions, err := s.ListVersionStrings()
	if err != nil {
		return ResolvedVersion{}, err
	}
	if len(versions) > 0 {
		v := versions[len(versions)-1]
		bin, binErr := s.BinaryPath(v)
		return ResolvedVersion{Version: v, BinaryPath: bin, Source: "latest"}, binErr
	}

	return ResolvedVersion{}, ErrNoVersion
}

// ResolveInstalledVersion resolves a version query against installed versions.
// Supports exact match, "latest"/"stable" aliases, and prefix matching.
func (s *Service) ResolveInstalledVersion(query string) (string, error) {
	installed, err := s.ListVersionStrings()
	if err != nil {
		return "", err
	}

	// Exact match
	for _, v := range installed {
		if v == query {
			return v, nil
		}
	}

	// latest/stable aliases
	if query == "latest" || query == "stable" {
		if len(installed) > 0 {
			return installed[len(installed)-1], nil
		}
		return "", &ActionableError{
			Err:        fmt.Errorf("no versions installed"),
			Suggestion: "gdt install latest",
		}
	}

	// Prefix match (e.g. "4.3" matches "4.3.1")
	for i := len(installed) - 1; i >= 0; i-- {
		if strings.HasPrefix(installed[i], query) {
			return installed[i], nil
		}
	}

	return "", &ActionableError{
		Err:        fmt.Errorf("version %q not found", query),
		Suggestion: fmt.Sprintf("gdt install %s", query),
	}
}

// resolveFromFile walks parent directories from startDir looking for .godot-version.
func resolveFromFile(startDir string) (string, error) {
	dir := startDir
	for {
		path := filepath.Join(dir, ".godot-version")
		data, err := os.ReadFile(path)
		if err == nil {
			v := strings.TrimSpace(string(data))
			if v != "" {
				return v, nil
			}
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return "", fmt.Errorf("no .godot-version file found")
}
