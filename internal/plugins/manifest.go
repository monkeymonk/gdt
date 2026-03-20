package plugins

import "github.com/BurntSushi/toml"

// HookEvent represents a lifecycle event that plugins can hook into.
type HookEvent string

const (
	BeforeExport HookEvent = "before_export"
	AfterExport  HookEvent = "after_export"
	BeforeBuild  HookEvent = "before_build"
)

// Hooks defines optional shell commands to run at lifecycle events.
type Hooks struct {
	BeforeExport string `toml:"before_export"`
	AfterExport  string `toml:"after_export"`
	BeforeBuild  string `toml:"before_build"`
}

type Manifest struct {
	Name        string   `toml:"name"`
	Version     string   `toml:"version"`
	Commands    []string `toml:"commands"`
	RequiresGdt string   `toml:"requires_gdt"`
	Description string   `toml:"description"`
	Hooks       Hooks    `toml:"hooks"`
}

// HookFor returns the shell command for a given hook event, or empty string.
func (m *Manifest) HookFor(event HookEvent) string {
	switch event {
	case BeforeExport:
		return m.Hooks.BeforeExport
	case AfterExport:
		return m.Hooks.AfterExport
	case BeforeBuild:
		return m.Hooks.BeforeBuild
	}
	return ""
}

func ParseManifest(data []byte) (*Manifest, error) {
	var m Manifest
	if err := toml.Unmarshal(data, &m); err != nil {
		return nil, err
	}
	return &m, nil
}
