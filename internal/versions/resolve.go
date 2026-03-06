package versions

import (
	"errors"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

var ErrNoVersion = errors.New("no godot version found\n\n  Set one with: gdt use <version>\n  Or pin a project: gdt local <version>")

func Resolve(startDir string, envVersion string, globalDefault string, installed []string) (string, error) {
	// 1. .godot-version in current or parent dirs
	if v := findVersionFile(startDir); v != "" {
		return v, nil
	}

	// 2. Environment variable
	if envVersion != "" {
		return envVersion, nil
	}

	// 3. Global default
	if globalDefault != "" {
		return globalDefault, nil
	}

	// 4. Latest installed
	if len(installed) > 0 {
		sorted := make([]string, len(installed))
		copy(sorted, installed)
		sort.Strings(sorted)
		return sorted[len(sorted)-1], nil
	}

	return "", ErrNoVersion
}

func findVersionFile(startDir string) string {
	dir := startDir
	for {
		path := filepath.Join(dir, ".godot-version")
		data, err := os.ReadFile(path)
		if err == nil {
			v := strings.TrimSpace(string(data))
			if v != "" {
				return v
			}
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return ""
		}
		dir = parent
	}
}
