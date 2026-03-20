package project

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParsePresetsEmpty(t *testing.T) {
	dir := t.TempDir()

	_, err := ParsePresets(dir)
	if err == nil {
		t.Error("should error when no export_presets.cfg")
	}
}

func TestParsePresets(t *testing.T) {
	dir := t.TempDir()
	content := `[preset.0]

name="Linux/X11"
platform="Linux/X11"

[preset.0.options]

[preset.1]

name="Windows Desktop"
platform="Windows Desktop"

[preset.1.options]
`
	os.WriteFile(filepath.Join(dir, "export_presets.cfg"), []byte(content), 0644)

	presets, err := ParsePresets(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(presets) != 2 {
		t.Errorf("expected 2 presets, got %d", len(presets))
	}
	if presets[0] != "Linux/X11" {
		t.Errorf("preset[0] = %q, want %q", presets[0], "Linux/X11")
	}
	if presets[1] != "Windows Desktop" {
		t.Errorf("preset[1] = %q, want %q", presets[1], "Windows Desktop")
	}
}

func TestDefaultOutputDir(t *testing.T) {
	got := DefaultOutputDir("Linux/X11")
	want := filepath.Join("dist", "Linux-X11")
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
