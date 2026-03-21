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

func TestParseManifestV2(t *testing.T) {
	content := `
name = "dotnet"
version = "1.0.0"
protocol = 2
commands = ["dotnet"]
requires_gdt = ">=1.0"
description = "C# tooling"

[contributions]
templates = ["csharp-starter"]
presets = ["android-mono"]
ci_providers = ["github-dotnet"]
hooks = ["after_new", "before_export"]
doctor = true
completions = true
`
	m, err := ParseManifest([]byte(content))
	if err != nil {
		t.Fatal(err)
	}
	if m.Protocol != 2 {
		t.Errorf("protocol = %d, want 2", m.Protocol)
	}
	if len(m.Contributions.Templates) != 1 || m.Contributions.Templates[0] != "csharp-starter" {
		t.Errorf("templates = %v, want [csharp-starter]", m.Contributions.Templates)
	}
	if len(m.Contributions.Presets) != 1 {
		t.Errorf("presets = %v, want [android-mono]", m.Contributions.Presets)
	}
	if len(m.Contributions.CIProviders) != 1 {
		t.Errorf("ci_providers = %v, want [github-dotnet]", m.Contributions.CIProviders)
	}
	if len(m.Contributions.Hooks) != 2 {
		t.Errorf("hooks = %v, want [after_new, before_export]", m.Contributions.Hooks)
	}
	if !m.Contributions.Doctor {
		t.Error("doctor should be true")
	}
	if !m.Contributions.Completions {
		t.Error("completions should be true")
	}
}

func TestParseManifestV1Compat(t *testing.T) {
	content := `
name = "assets"
version = "1.0.0"
commands = ["assets"]
`
	m, err := ParseManifest([]byte(content))
	if err != nil {
		t.Fatal(err)
	}
	if m.Protocol != 0 {
		t.Errorf("protocol = %d, want 0 (unset)", m.Protocol)
	}
	if m.HasContributions() {
		t.Error("V1 manifest should not have contributions")
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
