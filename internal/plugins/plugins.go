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

// discover scans pluginsDir for installed plugins, returning nil, nil
// if pluginsDir does not exist. Stats pluginsDir first rather than
// relying solely on os.ReadDir's error classification: on Windows,
// ReadDir against a path that exists but is a regular file (not a
// directory) returns ERROR_PATH_NOT_FOUND, which os.IsNotExist also
// treats as "not exist" — silently misreporting a real problem (e.g.
// a corrupted plugins directory) as simply "no plugins installed".
// Stat distinguishes missing from wrong-type, on every platform.
func discover(pluginsDir string) ([]Plugin, error) {
	if _, err := os.Stat(pluginsDir); err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	entries, err := os.ReadDir(pluginsDir)
	if err != nil {
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
