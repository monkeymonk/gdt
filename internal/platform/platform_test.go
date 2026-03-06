package platform

import (
	"runtime"
	"testing"
)

func TestDetect(t *testing.T) {
	p := Detect()
	if p.OS == "" {
		t.Fatal("OS should not be empty")
	}
	if p.Arch == "" {
		t.Fatal("Arch should not be empty")
	}
}

func TestArtifactName(t *testing.T) {
	tests := []struct {
		name     string
		os       string
		arch     string
		mono     bool
		expected string
	}{
		{"linux amd64", "linux", "amd64", false, "linux.x86_64"},
		{"linux amd64 mono", "linux", "amd64", true, "linux.x86_64"},
		{"darwin amd64", "darwin", "amd64", false, "macos.universal"},
		{"darwin arm64", "darwin", "arm64", false, "macos.universal"},
		{"windows amd64", "windows", "amd64", false, "win64.exe"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := Info{OS: tt.os, Arch: tt.arch}
			got := p.ArtifactName()
			if got != tt.expected {
				t.Errorf("ArtifactName() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestArtifactUnsupported(t *testing.T) {
	p := Info{OS: "freebsd", Arch: "amd64"}
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for unsupported platform")
		}
	}()
	p.ArtifactName()
}

func TestDefaultHome(t *testing.T) {
	tests := []struct {
		name string
		os   string
	}{
		{"linux", "linux"},
		{"darwin", "darwin"},
		{"windows", "windows"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := Info{OS: tt.os, Arch: "amd64"}
			home := p.DefaultHome()
			if home == "" {
				t.Fatal("DefaultHome should not be empty")
			}
		})
	}
}

func TestDetectMatchesRuntime(t *testing.T) {
	p := Detect()
	if p.OS != runtime.GOOS {
		t.Errorf("OS = %q, want %q", p.OS, runtime.GOOS)
	}
	if p.Arch != runtime.GOARCH {
		t.Errorf("Arch = %q, want %q", p.Arch, runtime.GOARCH)
	}
}
