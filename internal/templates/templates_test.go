package templates

import (
	"os"
	"path/filepath"
	"testing"
)

func TestListEmpty(t *testing.T) {
	dir := t.TempDir()
	list, err := List(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 0 {
		t.Errorf("expected 0, got %d", len(list))
	}
}

func TestListTemplates(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "4.3"), 0755)
	os.MkdirAll(filepath.Join(dir, "4.2"), 0755)

	list, err := List(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 2 {
		t.Errorf("expected 2, got %d", len(list))
	}
}

func TestIsInstalled(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "4.3"), 0755)

	if !IsInstalled(dir, "4.3") {
		t.Error("should be installed")
	}
	if IsInstalled(dir, "4.2") {
		t.Error("should not be installed")
	}
}

func TestArtifactName(t *testing.T) {
	tests := []struct {
		version string
		mono    bool
		want    string
	}{
		{"4.3", false, "Godot_v4.3-stable_export_templates.tpz"},
		{"4.3", true, "Godot_v4.3-stable_mono_export_templates.tpz"},
	}
	for _, tt := range tests {
		t.Run(tt.version, func(t *testing.T) {
			got := ArtifactName(tt.version, tt.mono)
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}
