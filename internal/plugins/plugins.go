package plugins

import (
	"os"
	"path/filepath"
)

const ManifestFile = "plugin.toml"

type Plugin struct {
	Dir      string
	Manifest Manifest
}

func discover(pluginsDir string) ([]Plugin, error) {
	entries, err := os.ReadDir(pluginsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var plugins []Plugin
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		manifestPath := filepath.Join(pluginsDir, e.Name(), ManifestFile)
		data, err := os.ReadFile(manifestPath)
		if err != nil {
			continue
		}
		m, err := ParseManifest(data)
		if err != nil {
			continue
		}
		plugins = append(plugins, Plugin{
			Dir:      filepath.Join(pluginsDir, e.Name()),
			Manifest: *m,
		})
	}
	return plugins, nil
}

func findForCommand(plugins []Plugin, command string) (Plugin, bool) {
	for _, p := range plugins {
		for _, c := range p.Manifest.Commands {
			if c == command {
				return p, true
			}
		}
	}
	return Plugin{}, false
}
