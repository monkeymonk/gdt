package project

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func ParsePresets(projectDir string) ([]string, error) {
	path := filepath.Join(projectDir, "export_presets.cfg")
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("no export presets found\n\n  Configure them in the Godot editor: Project > Export")
	}
	defer f.Close()

	var presets []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "name=\"") {
			name := strings.TrimPrefix(line, "name=\"")
			name = strings.TrimSuffix(name, "\"")
			presets = append(presets, name)
		}
	}

	return presets, nil
}

func DefaultOutputDir(preset string) string {
	safe := strings.ReplaceAll(preset, "/", "-")
	safe = strings.ReplaceAll(safe, " ", "-")
	return filepath.Join("dist", safe)
}
