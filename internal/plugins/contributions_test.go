package plugins

import (
	"os"
	"path/filepath"
	"testing"
)

func writeV2Plugin(t *testing.T, dir, name string, manifest string) string {
	t.Helper()
	pluginDir := filepath.Join(dir, "gdt-"+name)
	os.MkdirAll(pluginDir, 0755)
	os.WriteFile(filepath.Join(pluginDir, ManifestFile), []byte(manifest), 0644)
	return pluginDir
}

func TestDiscoverTemplates(t *testing.T) {
	dir := t.TempDir()
	pdir := writeV2Plugin(t, dir, "starter", `
name = "starter"
version = "1.0.0"
protocol = 2

[contributions]
templates = ["fps", "rpg"]
`)
	// Create template directories
	os.MkdirAll(filepath.Join(pdir, "templates", "fps"), 0755)
	os.WriteFile(filepath.Join(pdir, "templates", "fps", "project.godot"), []byte(""), 0644)
	os.MkdirAll(filepath.Join(pdir, "templates", "rpg"), 0755)
	os.WriteFile(filepath.Join(pdir, "templates", "rpg", "project.godot"), []byte(""), 0644)

	svc := NewService(dir)
	templates, err := svc.DiscoverTemplates()
	if err != nil {
		t.Fatal(err)
	}
	if len(templates) != 2 {
		t.Fatalf("expected 2 templates, got %d", len(templates))
	}
}

func TestDiscoverPresets(t *testing.T) {
	dir := t.TempDir()
	pdir := writeV2Plugin(t, dir, "mobile", `
name = "mobile"
version = "1.0.0"
protocol = 2

[contributions]
presets = ["android"]
`)
	os.MkdirAll(filepath.Join(pdir, "presets"), 0755)
	os.WriteFile(filepath.Join(pdir, "presets", "android.cfg"), []byte("[preset]\nname=\"Android\""), 0644)

	svc := NewService(dir)
	presets, err := svc.DiscoverPresets()
	if err != nil {
		t.Fatal(err)
	}
	if len(presets) != 1 {
		t.Fatalf("expected 1 preset, got %d", len(presets))
	}
	if presets[0].Name != "android" {
		t.Errorf("preset name = %q, want android", presets[0].Name)
	}
}

func TestDiscoverCIProviders(t *testing.T) {
	dir := t.TempDir()
	pdir := writeV2Plugin(t, dir, "citools", `
name = "citools"
version = "1.0.0"
protocol = 2

[contributions]
ci_providers = ["bitbucket"]
`)
	os.MkdirAll(filepath.Join(pdir, "ci"), 0755)
	os.WriteFile(filepath.Join(pdir, "ci", "bitbucket.yml"), []byte("pipeline:"), 0644)

	svc := NewService(dir)
	providers, err := svc.DiscoverCIProviders()
	if err != nil {
		t.Fatal(err)
	}
	if len(providers) != 1 {
		t.Fatalf("expected 1 provider, got %d\n", len(providers))
	}
}

func TestDiscoverTemplates_MissingDir_Warns(t *testing.T) {
	dir := t.TempDir()
	writeV2Plugin(t, dir, "broken", `
name = "broken"
version = "1.0.0"
protocol = 2

[contributions]
templates = ["missing"]
`)
	// Don't create the templates/missing directory

	svc := NewService(dir)
	templates, err := svc.DiscoverTemplates()
	if err != nil {
		t.Fatal(err)
	}
	// Missing template directories are skipped with a warning, not returned
	if len(templates) != 0 {
		t.Errorf("expected 0 valid templates, got %d", len(templates))
	}
}

func TestDiscoverDoctorPlugins(t *testing.T) {
	dir := t.TempDir()
	writeV2Plugin(t, dir, "checker", `
name = "checker"
version = "1.0.0"
protocol = 2

[contributions]
doctor = true
`)

	svc := NewService(dir)
	doctors := svc.DiscoverDoctorPlugins()
	if len(doctors) != 1 {
		t.Fatalf("expected 1 doctor plugin, got %d", len(doctors))
	}
}
