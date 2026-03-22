package metadata

import (
	"testing"

	"github.com/monkeymonk/gdt/internal/platform"
)

func TestArtifactName(t *testing.T) {
	tests := []struct {
		name     string
		plat     platform.Info
		version  string
		mono     bool
		expected string
	}{
		{"linux amd64", platform.Info{OS: "linux", Arch: "amd64"}, "4.3", false, "Godot_v4.3-stable_linux.x86_64.zip"},
		{"linux amd64 mono", platform.Info{OS: "linux", Arch: "amd64"}, "4.3", true, "Godot_v4.3-stable_mono_linux.x86_64.zip"},
		{"darwin amd64", platform.Info{OS: "darwin", Arch: "amd64"}, "4.3", false, "Godot_v4.3-stable_macos.universal.zip"},
		{"darwin arm64", platform.Info{OS: "darwin", Arch: "arm64"}, "4.3", false, "Godot_v4.3-stable_macos.universal.zip"},
		{"windows amd64", platform.Info{OS: "windows", Arch: "amd64"}, "4.3", false, "Godot_v4.3-stable_win64.exe.zip"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ArtifactName(tt.plat, tt.version, tt.mono)
			if err != nil {
				t.Fatalf("ArtifactName() unexpected error: %v", err)
			}
			if got != tt.expected {
				t.Errorf("ArtifactName() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestArtifactNameUnsupported(t *testing.T) {
	_, err := ArtifactName(platform.Info{OS: "freebsd", Arch: "amd64"}, "4.3", false)
	if err == nil {
		t.Error("expected error for unsupported platform")
	}
}

func TestTemplateArtifactName(t *testing.T) {
	tests := []struct {
		name     string
		version  string
		mono     bool
		expected string
	}{
		{"standard", "4.3", false, "Godot_v4.3-stable_export_templates.tpz"},
		{"mono", "4.3", true, "Godot_v4.3-stable_mono_export_templates.tpz"},
		{"with patch", "4.2.1", false, "Godot_v4.2.1-stable_export_templates.tpz"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := TemplateArtifactName(tt.version, tt.mono)
			if got != tt.expected {
				t.Errorf("TemplateArtifactName() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestResolveEngineArtifact(t *testing.T) {
	release := &Release{
		Version: "4.3",
		Assets: map[string]string{
			"Godot_v4.3-stable_linux.x86_64.zip":          "https://example.com/linux.zip",
			"Godot_v4.3-stable_macos.universal.zip":       "https://example.com/macos.zip",
			"Godot_v4.3-stable_win64.exe.zip":             "https://example.com/win.zip",
			"Godot_v4.3-stable_mono_linux.x86_64.zip":     "https://example.com/mono_linux.zip",
			"Godot_v4.3-stable_export_templates.tpz":      "https://example.com/templates.tpz",
			"Godot_v4.3-stable_mono_export_templates.tpz": "https://example.com/mono_templates.tpz",
		},
	}

	tests := []struct {
		name     string
		plat     platform.Info
		mono     bool
		expected string
	}{
		{"linux standard", platform.Info{OS: "linux", Arch: "amd64"}, false, "Godot_v4.3-stable_linux.x86_64.zip"},
		{"linux mono", platform.Info{OS: "linux", Arch: "amd64"}, true, "Godot_v4.3-stable_mono_linux.x86_64.zip"},
		{"macos standard", platform.Info{OS: "darwin", Arch: "arm64"}, false, "Godot_v4.3-stable_macos.universal.zip"},
		{"windows standard", platform.Info{OS: "windows", Arch: "amd64"}, false, "Godot_v4.3-stable_win64.exe.zip"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ResolveEngineArtifact(release, tt.plat, tt.mono)
			if err != nil {
				t.Fatalf("ResolveEngineArtifact() unexpected error: %v", err)
			}
			if got != tt.expected {
				t.Errorf("ResolveEngineArtifact() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestResolveEngineArtifactUnsupportedPlatform(t *testing.T) {
	release := &Release{Version: "4.3", Assets: map[string]string{}}
	_, err := ResolveEngineArtifact(release, platform.Info{OS: "freebsd", Arch: "amd64"}, false)
	if err == nil {
		t.Error("expected error for unsupported platform")
	}
}

func TestResolveEngineArtifactFallback(t *testing.T) {
	// When asset not in map, should return constructed name
	release := &Release{Version: "4.3", Assets: map[string]string{}}
	got, err := ResolveEngineArtifact(release, platform.Info{OS: "linux", Arch: "amd64"}, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := "Godot_v4.3-stable_linux.x86_64.zip"
	if got != expected {
		t.Errorf("got %q, want %q", got, expected)
	}
}
