package plugins

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Remove deletes an installed plugin by directory name, repo slug, or manifest name.
func (s *Service) Remove(name string) error {
	// Try exact directory name first
	dir := filepath.Join(s.Dir, name)
	if _, err := os.Stat(dir); err == nil {
		return os.RemoveAll(dir)
	}

	// Try extracting last component from owner/repo slug
	if strings.Contains(name, "/") {
		parts := strings.Split(strings.TrimSuffix(name, "/"), "/")
		dir = filepath.Join(s.Dir, parts[len(parts)-1])
		if _, err := os.Stat(dir); err == nil {
			return os.RemoveAll(dir)
		}
	}

	// Try matching by manifest name
	plugins, err := discover(s.Dir)
	if err == nil {
		for _, p := range plugins {
			if p.Manifest.Name == name {
				return os.RemoveAll(p.Dir)
			}
		}
	}

	return fmt.Errorf("plugin %s not found", name)
}
