package plugins

import (
	"fmt"
	"os"
	"path/filepath"
)

// Scaffold creates a new plugin directory with a manifest and README.
// The directory is created at ./gdt-<name> relative to the current working directory.
func (s *Service) Scaffold(name string) (string, error) {
	dir := filepath.Join(".", "gdt-"+name)

	if _, err := os.Stat(dir); err == nil {
		return "", fmt.Errorf("directory %s already exists", dir)
	}

	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}

	manifest := fmt.Sprintf(`name = %q
version = "0.1.0"
commands = [%q]
requires_gdt = ">=1.0"
description = ""
`, name, name)

	if err := os.WriteFile(filepath.Join(dir, ManifestFile), []byte(manifest), 0644); err != nil {
		return "", err
	}

	readme := fmt.Sprintf("# gdt-%s\n\nA gdt plugin.\n\n## Usage\n\n```sh\ngdt %s\n```\n", name, name)
	if err := os.WriteFile(filepath.Join(dir, "README.md"), []byte(readme), 0644); err != nil {
		return "", err
	}

	return dir, nil
}
