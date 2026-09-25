package engine

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Resolve determines which Godot version to use for the current working
// directory (startDir) by checking sources in order of precedence:
// 1. .godot-version file (walks parent dirs from startDir)
// 2. GDT_GODOT_VERSION environment variable
// 3. Config default version
// 4. Latest installed version
// This is the standard resolution order used when no explicit version
// is specified. Use Resolve when the caller needs to determine the version
// for a given directory per the full precedence chain.
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
		v := versions[0]
		bin, binErr := s.BinaryPath(v)
		return ResolvedVersion{Version: v, BinaryPath: bin, Source: "latest"}, binErr
	}

	return ResolvedVersion{}, ErrNoVersion
}

// ResolveInstalledVersion matches a version query against already-installed
// versions only. It supports three query types:
// 1. Exact version match (e.g., "4.3.1")
// 2. "latest" or "stable" aliases (returns the newest installed version)
// 3. Prefix match (e.g., "4.3" matches "4.3.1", returns the highest match)
// Unlike Resolve, this function does not check .godot-version files, environment
// variables, or config defaults — only what is physically installed in the
// versions directory. Use ResolveInstalledVersion when validating user input
// against the set of available installations.
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
			return installed[0], nil
		}
		return "", Actionable(
			fmt.Errorf("no versions installed"),
			"gdt install latest",
		)
	}

	// Prefix match (e.g. "4.3" matches "4.3.1") — installed is newest
	// first, so the first match found is the highest matching version.
	for i := 0; i < len(installed); i++ {
		if strings.HasPrefix(installed[i], query) {
			return installed[i], nil
		}
	}

	return "", Actionable(
		fmt.Errorf("version %q not found", query),
		fmt.Sprintf("gdt install %s", query),
	)
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
