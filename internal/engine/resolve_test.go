package engine

import (
	"os"
	"path/filepath"
	"testing"
)

// --- resolveFromFile tests ---

func TestResolveFromFile_Found(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, ".godot-version"), []byte("4.2.1\n"), 0o644)

	v, err := resolveFromFile(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v != "4.2.1" {
		t.Errorf("expected 4.2.1, got %s", v)
	}
}

func TestResolveFromFile_ParentDir(t *testing.T) {
	parent := t.TempDir()
	child := filepath.Join(parent, "sub", "deep")
	os.MkdirAll(child, 0o755)
	os.WriteFile(filepath.Join(parent, ".godot-version"), []byte("4.3.0"), 0o644)

	v, err := resolveFromFile(child)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v != "4.3.0" {
		t.Errorf("expected 4.3.0, got %s", v)
	}
}

func TestResolveFromFile_NotFound(t *testing.T) {
	dir := t.TempDir()
	_, err := resolveFromFile(dir)
	if err == nil {
		t.Fatal("expected error when no .godot-version exists")
	}
}

func TestResolveFromFile_EmptyFile(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, ".godot-version"), []byte("  \n"), 0o644)

	_, err := resolveFromFile(dir)
	if err == nil {
		t.Fatal("expected error for empty .godot-version")
	}
}

// --- Resolve tests ---

func TestResolve_FromFile(t *testing.T) {
	svc := testService(t)
	setupFakeVersion(t, svc, "4.2.1")

	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, ".godot-version"), []byte("4.2.1"), 0o644)

	rv, err := svc.Resolve(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rv.Version != "4.2.1" {
		t.Errorf("expected version 4.2.1, got %s", rv.Version)
	}
	if rv.Source != "file" {
		t.Errorf("expected source file, got %s", rv.Source)
	}
}

func TestResolve_FromEnv(t *testing.T) {
	svc := testService(t)
	setupFakeVersion(t, svc, "4.3.0")

	t.Setenv("GDT_GODOT_VERSION", "4.3.0")

	dir := t.TempDir() // no .godot-version here
	rv, err := svc.Resolve(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rv.Version != "4.3.0" {
		t.Errorf("expected version 4.3.0, got %s", rv.Version)
	}
	if rv.Source != "env" {
		t.Errorf("expected source env, got %s", rv.Source)
	}
}

func TestResolve_FromConfig(t *testing.T) {
	svc := testService(t)
	svc.Config.DefaultVersion = "4.2.1"
	setupFakeVersion(t, svc, "4.2.1")

	t.Setenv("GDT_GODOT_VERSION", "")

	dir := t.TempDir()
	rv, err := svc.Resolve(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rv.Version != "4.2.1" {
		t.Errorf("expected version 4.2.1, got %s", rv.Version)
	}
	if rv.Source != "config" {
		t.Errorf("expected source config, got %s", rv.Source)
	}
}

func TestResolve_LatestInstalled(t *testing.T) {
	svc := testService(t)
	setupFakeVersion(t, svc, "4.1.0")
	setupFakeVersion(t, svc, "4.3.0")
	setupFakeVersion(t, svc, "4.2.1")

	t.Setenv("GDT_GODOT_VERSION", "")

	dir := t.TempDir()
	rv, err := svc.Resolve(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rv.Version != "4.3.0" {
		t.Errorf("expected latest version 4.3.0, got %s", rv.Version)
	}
	if rv.Source != "latest" {
		t.Errorf("expected source latest, got %s", rv.Source)
	}
}

func TestResolve_NoVersion(t *testing.T) {
	svc := testService(t)

	t.Setenv("GDT_GODOT_VERSION", "")

	dir := t.TempDir()
	_, err := svc.Resolve(dir)
	if err != ErrNoVersion {
		t.Errorf("expected ErrNoVersion, got %v", err)
	}
}

func TestResolve_FileTakesPrecedence(t *testing.T) {
	svc := testService(t)
	svc.Config.DefaultVersion = "4.1.0"
	setupFakeVersion(t, svc, "4.1.0")
	setupFakeVersion(t, svc, "4.2.1")

	t.Setenv("GDT_GODOT_VERSION", "4.1.0")

	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, ".godot-version"), []byte("4.2.1"), 0o644)

	rv, err := svc.Resolve(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rv.Version != "4.2.1" {
		t.Errorf("expected file version 4.2.1, got %s", rv.Version)
	}
	if rv.Source != "file" {
		t.Errorf("expected source file, got %s", rv.Source)
	}
}

// --- ResolveProject tests ---

func TestResolveProject(t *testing.T) {
	svc := testService(t)
	setupFakeVersion(t, svc, "4.2.1")

	projectDir := t.TempDir()
	os.WriteFile(filepath.Join(projectDir, "project.godot"), []byte(""), 0o644)
	os.WriteFile(filepath.Join(projectDir, ".godot-version"), []byte("4.2.1"), 0o644)

	root, rv, err := svc.ResolveProject(projectDir)
	if err != nil {
		t.Fatal(err)
	}
	if root != projectDir {
		t.Errorf("expected root %s, got %s", projectDir, root)
	}
	if rv.Version != "4.2.1" {
		t.Errorf("expected version 4.2.1, got %s", rv.Version)
	}
	if rv.Source != "file" {
		t.Errorf("expected source file, got %s", rv.Source)
	}
}

// --- ResolveInstalledVersion tests ---

func TestResolveInstalledVersion_ExactMatch(t *testing.T) {
	svc := testService(t)
	setupFakeVersion(t, svc, "4.2.1")
	setupFakeVersion(t, svc, "4.3.0")

	v, err := svc.ResolveInstalledVersion("4.2.1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v != "4.2.1" {
		t.Errorf("expected 4.2.1, got %s", v)
	}
}

func TestResolveInstalledVersion_Latest(t *testing.T) {
	svc := testService(t)
	setupFakeVersion(t, svc, "4.1.0")
	setupFakeVersion(t, svc, "4.3.0")

	v, err := svc.ResolveInstalledVersion("latest")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v != "4.3.0" {
		t.Errorf("expected 4.3.0, got %s", v)
	}
}

func TestResolveInstalledVersion_Stable(t *testing.T) {
	svc := testService(t)
	setupFakeVersion(t, svc, "4.1.0")
	setupFakeVersion(t, svc, "4.2.1")

	v, err := svc.ResolveInstalledVersion("stable")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v != "4.2.1" {
		t.Errorf("expected 4.2.1, got %s", v)
	}
}

func TestResolveInstalledVersion_PrefixMatch(t *testing.T) {
	svc := testService(t)
	setupFakeVersion(t, svc, "4.3.0")
	setupFakeVersion(t, svc, "4.3.1")
	setupFakeVersion(t, svc, "4.2.1")

	v, err := svc.ResolveInstalledVersion("4.3")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v != "4.3.1" {
		t.Errorf("expected 4.3.1 (latest prefix match), got %s", v)
	}
}

func TestResolveInstalledVersion_NotFound(t *testing.T) {
	svc := testService(t)
	setupFakeVersion(t, svc, "4.2.1")

	_, err := svc.ResolveInstalledVersion("4.9")
	if err == nil {
		t.Fatal("expected error for unmatched version")
	}
	ae, ok := err.(*ActionableError)
	if !ok {
		t.Fatalf("expected ActionableError, got %T", err)
	}
	if ae.Suggestion != "gdt install 4.9" {
		t.Errorf("unexpected suggestion: %s", ae.Suggestion)
	}
}

func TestResolveInstalledVersion_LatestEmpty(t *testing.T) {
	svc := testService(t)

	_, err := svc.ResolveInstalledVersion("latest")
	if err == nil {
		t.Fatal("expected error when no versions installed")
	}
	ae, ok := err.(*ActionableError)
	if !ok {
		t.Fatalf("expected ActionableError, got %T", err)
	}
	if ae.Suggestion != "gdt install latest" {
		t.Errorf("unexpected suggestion: %s", ae.Suggestion)
	}
}
