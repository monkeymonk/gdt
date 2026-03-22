package engine

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/monkeymonk/gdt/internal/config"
	"github.com/monkeymonk/gdt/internal/platform"
)

func testService(t *testing.T) *Service {
	t.Helper()
	home := t.TempDir()
	os.MkdirAll(filepath.Join(home, "versions"), 0o755)
	os.MkdirAll(filepath.Join(home, "templates"), 0o755)
	return NewService(home, platform.Info{OS: "linux", Arch: "amd64"}, &config.Config{})
}

func setupFakeVersion(t *testing.T, svc *Service, version string) {
	t.Helper()
	dir := filepath.Join(svc.VersionsDir(), version)
	os.MkdirAll(dir, 0o755)
	os.WriteFile(filepath.Join(dir, "Godot_v"+version+"-stable_linux.x86_64"), []byte("bin"), 0o755)
}

func TestList_Empty(t *testing.T) {
	svc := testService(t)
	versions, err := svc.List()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(versions) != 0 {
		t.Fatalf("expected 0 versions, got %d", len(versions))
	}
}

func TestList_MultipleVersions(t *testing.T) {
	svc := testService(t)
	setupFakeVersion(t, svc, "4.2.1")
	setupFakeVersion(t, svc, "4.3.0")
	setupFakeVersion(t, svc, "4.1.0")

	versions, err := svc.List()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(versions) != 3 {
		t.Fatalf("expected 3 versions, got %d", len(versions))
	}
	// Should be sorted alphabetically
	if versions[0].Version != "4.1.0" {
		t.Errorf("expected first version 4.1.0, got %s", versions[0].Version)
	}
	if versions[1].Version != "4.2.1" {
		t.Errorf("expected second version 4.2.1, got %s", versions[1].Version)
	}
	if versions[2].Version != "4.3.0" {
		t.Errorf("expected third version 4.3.0, got %s", versions[2].Version)
	}
}

func TestList_MarksDefault(t *testing.T) {
	svc := testService(t)
	svc.Config.DefaultVersion = "4.2.1"
	setupFakeVersion(t, svc, "4.2.1")
	setupFakeVersion(t, svc, "4.3.0")

	versions, err := svc.List()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, v := range versions {
		if v.Version == "4.2.1" && !v.IsDefault {
			t.Error("expected 4.2.1 to be marked as default")
		}
		if v.Version == "4.3.0" && v.IsDefault {
			t.Error("expected 4.3.0 to not be marked as default")
		}
	}
}

func TestListVersionStrings(t *testing.T) {
	svc := testService(t)
	setupFakeVersion(t, svc, "4.2.1")
	setupFakeVersion(t, svc, "4.3.0")

	strs, err := svc.ListVersionStrings()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(strs) != 2 {
		t.Fatalf("expected 2 strings, got %d", len(strs))
	}
	if strs[0] != "4.2.1" || strs[1] != "4.3.0" {
		t.Errorf("unexpected strings: %v", strs)
	}
}

func TestIsInstalled_True(t *testing.T) {
	svc := testService(t)
	setupFakeVersion(t, svc, "4.2.1")
	if !svc.IsInstalled("4.2.1") {
		t.Error("expected version to be installed")
	}
}

func TestIsInstalled_False(t *testing.T) {
	svc := testService(t)
	if svc.IsInstalled("4.9.9") {
		t.Error("expected version to not be installed")
	}
}

func TestBinaryPath_Canonical(t *testing.T) {
	svc := testService(t)
	dir := filepath.Join(svc.VersionsDir(), "4.2.1")
	os.MkdirAll(dir, 0o755)
	os.WriteFile(filepath.Join(dir, "godot"), []byte("bin"), 0o755)

	p, err := svc.BinaryPath("4.2.1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if filepath.Base(p) != "godot" {
		t.Errorf("expected canonical name godot, got %s", filepath.Base(p))
	}
}

func TestBinaryPath_ScanPattern(t *testing.T) {
	svc := testService(t)
	setupFakeVersion(t, svc, "4.2.1")

	p, err := svc.BinaryPath("4.2.1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if filepath.Base(p) != "Godot_v4.2.1-stable_linux.x86_64" {
		t.Errorf("expected scan pattern match, got %s", filepath.Base(p))
	}
}

func TestBinaryPath_NotInstalled(t *testing.T) {
	svc := testService(t)
	_, err := svc.BinaryPath("9.9.9")
	if err == nil {
		t.Fatal("expected error for missing version")
	}
	ae, ok := err.(*ActionableError)
	if !ok {
		t.Fatalf("expected ActionableError, got %T", err)
	}
	if ae.Suggestion != "gdt install 9.9.9" {
		t.Errorf("unexpected suggestion: %s", ae.Suggestion)
	}
}

func TestBinaryPath_NoBinary(t *testing.T) {
	svc := testService(t)
	// Create version dir but no binary
	os.MkdirAll(filepath.Join(svc.VersionsDir(), "4.2.1"), 0o755)

	_, err := svc.BinaryPath("4.2.1")
	if err == nil {
		t.Fatal("expected error for missing binary")
	}
	ae, ok := err.(*ActionableError)
	if !ok {
		t.Fatalf("expected ActionableError, got %T", err)
	}
	if ae.Suggestion != "gdt install 4.2.1 --force" {
		t.Errorf("unexpected suggestion: %s", ae.Suggestion)
	}
}

func TestBinaryPath_SkipsConsoleOnWindows(t *testing.T) {
	svc := testService(t)
	svc.Platform.OS = "windows"
	dir := filepath.Join(svc.VersionsDir(), "4.2.1")
	os.MkdirAll(dir, 0o755)
	// Create console variant first (alphabetically first)
	os.WriteFile(filepath.Join(dir, "Godot_v4.2.1-stable_console.exe"), []byte("bin"), 0o755)
	os.WriteFile(filepath.Join(dir, "Godot_v4.2.1-stable_win64.exe"), []byte("bin"), 0o755)

	p, err := svc.BinaryPath("4.2.1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if filepath.Base(p) != "Godot_v4.2.1-stable_win64.exe" {
		t.Errorf("expected non-console binary, got %s", filepath.Base(p))
	}
}
