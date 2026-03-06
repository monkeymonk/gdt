package versions

import (
	"os"
	"path/filepath"
	"testing"
)

func TestListEmpty(t *testing.T) {
	dir := t.TempDir()
	versions, err := List(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(versions) != 0 {
		t.Errorf("expected 0 versions, got %d", len(versions))
	}
}

func TestListVersions(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "4.3"), 0755)
	os.MkdirAll(filepath.Join(dir, "4.2"), 0755)
	os.MkdirAll(filepath.Join(dir, "4.3-mono"), 0755)

	versions, err := List(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(versions) != 3 {
		t.Errorf("expected 3 versions, got %d", len(versions))
	}
}

func TestIsInstalled(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "4.3"), 0755)

	if !IsInstalled(dir, "4.3") {
		t.Error("4.3 should be installed")
	}
	if IsInstalled(dir, "4.2") {
		t.Error("4.2 should not be installed")
	}
}

func TestRemove(t *testing.T) {
	dir := t.TempDir()
	vdir := filepath.Join(dir, "4.3")
	os.MkdirAll(vdir, 0755)
	os.WriteFile(filepath.Join(vdir, "godot"), []byte("binary"), 0755)

	err := Remove(dir, "4.3")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(vdir); !os.IsNotExist(err) {
		t.Error("version directory should be removed")
	}
}

func TestRemoveNotInstalled(t *testing.T) {
	dir := t.TempDir()
	err := Remove(dir, "4.3")
	if err == nil {
		t.Error("should error when removing non-installed version")
	}
}

func TestBinaryPath(t *testing.T) {
	tests := []struct {
		name    string
		goos    string
		version string
		want    string
	}{
		{"linux", "linux", "4.3", "4.3/godot"},
		{"windows", "windows", "4.3", "4.3/godot.exe"},
		{"mono", "linux", "4.3-mono", "4.3-mono/godot"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := BinaryPath(tt.version, tt.goos)
			if got != tt.want {
				t.Errorf("BinaryPath() = %q, want %q", got, tt.want)
			}
		})
	}
}
