package templates

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

func List(templatesDir string) ([]string, error) {
	entries, err := os.ReadDir(templatesDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var templates []string
	for _, e := range entries {
		if e.IsDir() {
			templates = append(templates, e.Name())
		}
	}
	sort.Strings(templates)
	return templates, nil
}

func IsInstalled(templatesDir string, version string) bool {
	_, err := os.Stat(filepath.Join(templatesDir, version))
	return err == nil
}

func ArtifactName(version string, mono bool) string {
	if mono {
		return fmt.Sprintf("Godot_v%s-stable_mono_export_templates.tpz", version)
	}
	return fmt.Sprintf("Godot_v%s-stable_export_templates.tpz", version)
}
