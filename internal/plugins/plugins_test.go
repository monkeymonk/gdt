package plugins

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseManifest(t *testing.T) {
	content := `
name = "assets"
version = "1.0.0"
commands = ["assets"]
requires_gdt = ">=1.0"
description = "Asset utilities"
`
	m, err := ParseManifest([]byte(content))
	if err != nil {
		t.Fatal(err)
	}
	if m.Name != "assets" {
		t.Errorf("name = %q, want %q", m.Name, "assets")
	}
	if len(m.Commands) != 1 || m.Commands[0] != "assets" {
		t.Errorf("commands = %v, want [assets]", m.Commands)
	}
	if m.Version != "1.0.0" {
		t.Errorf("version = %q, want %q", m.Version, "1.0.0")
	}
}

func TestParseManifestInvalid(t *testing.T) {
	_, err := ParseManifest([]byte(`[invalid`))
	if err == nil {
		t.Error("should error on invalid TOML")
	}
}

func TestDiscoverPlugins(t *testing.T) {
	dir := t.TempDir()

	pdir := filepath.Join(dir, "gdt-assets")
	os.MkdirAll(pdir, 0755)
	os.WriteFile(filepath.Join(pdir, "plugin.toml"), []byte(`
name = "assets"
version = "1.0.0"
commands = ["assets"]
requires_gdt = ">=1.0"
`), 0644)

	os.MkdirAll(filepath.Join(dir, "not-a-plugin"), 0755)

	plugins, err := Discover(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(plugins) != 1 {
		t.Errorf("expected 1 plugin, got %d", len(plugins))
	}
	if plugins[0].Manifest.Name != "assets" {
		t.Errorf("name = %q, want %q", plugins[0].Manifest.Name, "assets")
	}
}

func TestDiscoverEmpty(t *testing.T) {
	dir := t.TempDir()
	plugins, err := Discover(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(plugins) != 0 {
		t.Errorf("expected 0, got %d", len(plugins))
	}
}

func TestFindPluginForCommand(t *testing.T) {
	plugins := []Plugin{
		{
			Dir:      "/plugins/gdt-assets",
			Manifest: Manifest{Name: "assets", Commands: []string{"assets"}},
		},
		{
			Dir:      "/plugins/gdt-git",
			Manifest: Manifest{Name: "git", Commands: []string{"git"}},
		},
	}

	p, ok := FindForCommand(plugins, "assets")
	if !ok {
		t.Fatal("should find plugin for 'assets'")
	}
	if p.Manifest.Name != "assets" {
		t.Errorf("name = %q, want %q", p.Manifest.Name, "assets")
	}

	_, ok = FindForCommand(plugins, "unknown")
	if ok {
		t.Error("should not find plugin for 'unknown'")
	}
}
