package plugins

import (
	"log/slog"
	"os"
	"path/filepath"
	"strings"
)

// inPluginDir returns true if target is inside (or equal to) pluginDir.
func inPluginDir(pluginDir, target string) bool {
	clean := filepath.Clean(target)
	prefix := filepath.Clean(pluginDir) + string(filepath.Separator)
	return clean == filepath.Clean(pluginDir) || strings.HasPrefix(clean, prefix)
}

// PluginTemplate represents a template contributed by a plugin.
type PluginTemplate struct {
	Name       string // template name (e.g. "fps")
	PluginName string // owning plugin (e.g. "starter")
	Dir        string // absolute path to template directory
}

// PluginPreset represents an export preset contributed by a plugin.
type PluginPreset struct {
	Name       string // preset name (e.g. "android")
	PluginName string // owning plugin
	FilePath   string // absolute path to preset .cfg file
}

// PluginCIProvider represents a CI provider contributed by a plugin.
type PluginCIProvider struct {
	Name       string // provider name (e.g. "bitbucket")
	PluginName string // owning plugin
	FilePath   string // absolute path to CI config file
}

// DiscoverTemplates returns all templates contributed by installed plugins.
// Skips templates whose directories don't exist on disk.
func (s *Service) DiscoverTemplates() ([]PluginTemplate, error) {
	plugins, err := s.Discover()
	if err != nil {
		return nil, err
	}

	var templates []PluginTemplate
	for _, p := range plugins {
		for _, name := range p.Manifest.Contributions.Templates {
			dir := filepath.Join(p.Dir, "templates", name)
			if !inPluginDir(p.Dir, dir) {
				slog.Warn("plugin declares template with invalid path, skipping",
					"plugin", p.Manifest.Name, "template", name)
				continue
			}
			if _, err := os.Stat(dir); os.IsNotExist(err) {
				slog.Warn("plugin declares template but directory missing",
					"plugin", p.Manifest.Name, "template", name, "expected", dir)
				continue
			}
			templates = append(templates, PluginTemplate{
				Name:       name,
				PluginName: p.Manifest.Name,
				Dir:        dir,
			})
		}
	}
	return templates, nil
}

// DiscoverPresets returns all export presets contributed by installed plugins.
func (s *Service) DiscoverPresets() ([]PluginPreset, error) {
	plugins, err := s.Discover()
	if err != nil {
		return nil, err
	}

	var presets []PluginPreset
	for _, p := range plugins {
		for _, name := range p.Manifest.Contributions.Presets {
			fp := filepath.Join(p.Dir, "presets", name+".cfg")
			if !inPluginDir(p.Dir, fp) {
				slog.Warn("plugin declares preset with invalid path, skipping",
					"plugin", p.Manifest.Name, "preset", name)
				continue
			}
			if _, err := os.Stat(fp); os.IsNotExist(err) {
				slog.Warn("plugin declares preset but file missing",
					"plugin", p.Manifest.Name, "preset", name, "expected", fp)
				continue
			}
			presets = append(presets, PluginPreset{
				Name:       name,
				PluginName: p.Manifest.Name,
				FilePath:   fp,
			})
		}
	}
	return presets, nil
}

// DiscoverCIProviders returns all CI providers contributed by installed plugins.
func (s *Service) DiscoverCIProviders() ([]PluginCIProvider, error) {
	plugins, err := s.Discover()
	if err != nil {
		return nil, err
	}

	var providers []PluginCIProvider
	for _, p := range plugins {
		for _, name := range p.Manifest.Contributions.CIProviders {
			// Try common extensions
			var fp string
			for _, ext := range []string{".yml", ".yaml", ".sh"} {
				candidate := filepath.Join(p.Dir, "ci", name+ext)
				if !inPluginDir(p.Dir, candidate) {
					slog.Warn("plugin declares CI provider with invalid path, skipping",
						"plugin", p.Manifest.Name, "provider", name)
					break
				}
				if _, err := os.Stat(candidate); err == nil {
					fp = candidate
					break
				}
			}
			if fp == "" {
				slog.Warn("plugin declares CI provider but file missing",
					"plugin", p.Manifest.Name, "provider", name)
				continue
			}
			providers = append(providers, PluginCIProvider{
				Name:       name,
				PluginName: p.Manifest.Name,
				FilePath:   fp,
			})
		}
	}
	return providers, nil
}

// DiscoverDoctorPlugins returns all plugins that declare doctor = true.
func (s *Service) DiscoverDoctorPlugins() []Plugin {
	plugins, err := s.Discover()
	if err != nil {
		slog.Warn("failed to discover plugins for doctor checks", "error", err)
		return nil
	}
	var result []Plugin
	for _, p := range plugins {
		if p.Manifest.Contributions.Doctor {
			result = append(result, p)
		}
	}
	return result
}

// DiscoverCompletionPlugins returns all plugins that declare completions = true.
func (s *Service) DiscoverCompletionPlugins() []Plugin {
	plugins, err := s.Discover()
	if err != nil {
		slog.Warn("failed to discover plugins for completions", "error", err)
		return nil
	}
	var result []Plugin
	for _, p := range plugins {
		if p.Manifest.Contributions.Completions {
			result = append(result, p)
		}
	}
	return result
}
