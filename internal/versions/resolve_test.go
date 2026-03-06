package versions

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveFromGodotVersionFile(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, ".godot-version"), []byte("4.3\n"), 0644)

	v, err := Resolve(dir, "", "", nil)
	if err != nil {
		t.Fatal(err)
	}
	if v != "4.3" {
		t.Errorf("version = %q, want %q", v, "4.3")
	}
}

func TestResolveFromParentDir(t *testing.T) {
	dir := t.TempDir()
	sub := filepath.Join(dir, "scenes")
	os.MkdirAll(sub, 0755)
	os.WriteFile(filepath.Join(dir, ".godot-version"), []byte("4.2-mono"), 0644)

	v, err := Resolve(sub, "", "", nil)
	if err != nil {
		t.Fatal(err)
	}
	if v != "4.2-mono" {
		t.Errorf("version = %q, want %q", v, "4.2-mono")
	}
}

func TestResolveFromEnvVar(t *testing.T) {
	dir := t.TempDir()
	v, err := Resolve(dir, "4.1", "", nil)
	if err != nil {
		t.Fatal(err)
	}
	if v != "4.1" {
		t.Errorf("version = %q, want %q", v, "4.1")
	}
}

func TestResolveFromGlobalDefault(t *testing.T) {
	dir := t.TempDir()
	v, err := Resolve(dir, "", "4.0", nil)
	if err != nil {
		t.Fatal(err)
	}
	if v != "4.0" {
		t.Errorf("version = %q, want %q", v, "4.0")
	}
}

func TestResolveLatestInstalled(t *testing.T) {
	dir := t.TempDir()
	installed := []string{"4.1", "4.3", "4.2"}

	v, err := Resolve(dir, "", "", installed)
	if err != nil {
		t.Fatal(err)
	}
	if v != "4.3" {
		t.Errorf("version = %q, want %q", v, "4.3")
	}
}

func TestResolveNothingFound(t *testing.T) {
	dir := t.TempDir()
	_, err := Resolve(dir, "", "", nil)
	if err == nil {
		t.Error("should error when no version can be resolved")
	}
}

func TestResolvePriority(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, ".godot-version"), []byte("4.3"), 0644)

	v, err := Resolve(dir, "4.1", "4.0", []string{"3.5"})
	if err != nil {
		t.Fatal(err)
	}
	if v != "4.3" {
		t.Errorf("version = %q, want %q (file should win)", v, "4.3")
	}
}

func TestResolveTrimsWhitespace(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, ".godot-version"), []byte("  4.3  \n"), 0644)

	v, err := Resolve(dir, "", "", nil)
	if err != nil {
		t.Fatal(err)
	}
	if v != "4.3" {
		t.Errorf("version = %q, want %q", v, "4.3")
	}
}
