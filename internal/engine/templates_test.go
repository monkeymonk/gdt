package engine

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monkeymonk/gdt/internal/config"
	"github.com/monkeymonk/gdt/internal/platform"
)

func setupFakeTemplates(t *testing.T, svc *Service, version string) {
	t.Helper()
	dir := filepath.Join(svc.TemplatesDir(), version)
	os.MkdirAll(dir, 0o755)
	os.WriteFile(filepath.Join(dir, "dummy.tpz"), []byte("tpl"), 0o644)
}

func TestListTemplates_Empty(t *testing.T) {
	svc := testService(t)
	list, err := svc.ListTemplates()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(list) != 0 {
		t.Fatalf("expected 0 templates, got %d", len(list))
	}
}

func TestListTemplates_Populated(t *testing.T) {
	svc := testService(t)
	setupFakeTemplates(t, svc, "4.3.0")
	setupFakeTemplates(t, svc, "4.1.0")
	setupFakeTemplates(t, svc, "4.2.1")

	list, err := svc.ListTemplates()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(list) != 3 {
		t.Fatalf("expected 3 templates, got %d", len(list))
	}
	// Should be sorted
	if list[0] != "4.1.0" {
		t.Errorf("expected first 4.1.0, got %s", list[0])
	}
	if list[1] != "4.2.1" {
		t.Errorf("expected second 4.2.1, got %s", list[1])
	}
	if list[2] != "4.3.0" {
		t.Errorf("expected third 4.3.0, got %s", list[2])
	}
}

func TestListTemplates_NoDir(t *testing.T) {
	svc := testService(t)
	// Remove the templates dir created by testService
	os.RemoveAll(svc.TemplatesDir())

	list, err := svc.ListTemplates()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(list) != 0 {
		t.Fatalf("expected 0 templates, got %d", len(list))
	}
}

func TestTemplatesInstalled_True(t *testing.T) {
	svc := testService(t)
	setupFakeTemplates(t, svc, "4.2.1")
	if !svc.TemplatesInstalled("4.2.1") {
		t.Error("expected templates to be installed")
	}
}

func TestTemplatesInstalled_False(t *testing.T) {
	svc := testService(t)
	if svc.TemplatesInstalled("4.9.9") {
		t.Error("expected templates to not be installed")
	}
}

func TestRemoveTemplates_NotInstalled(t *testing.T) {
	svc := NewService(t.TempDir(), platform.Info{OS: "linux", Arch: "amd64"}, &config.Config{})
	err := svc.RemoveTemplates("4.3-stable")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var ae *ActionableError
	if !errors.As(err, &ae) {
		t.Fatalf("expected *ActionableError, got %T", err)
	}
	if ae.Suggestion != "gdt templates list" {
		t.Errorf("expected suggestion %q, got %q", "gdt templates list", ae.Suggestion)
	}
	if !strings.Contains(err.Error(), "4.3-stable") {
		t.Errorf("expected error message to contain %q, got %q", "4.3-stable", err.Error())
	}
}

func TestRemoveTemplates_Removes(t *testing.T) {
	home := t.TempDir()
	dir := filepath.Join(home, "templates", "4.3-stable")
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("failed to create dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "dummy.tpz"), []byte("tpl"), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	svc := NewService(home, platform.Info{OS: "linux", Arch: "amd64"}, &config.Config{})
	if err := svc.RemoveTemplates("4.3-stable"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := os.Stat(dir); !os.IsNotExist(err) {
		t.Errorf("expected dir to no longer exist, stat err: %v", err)
	}
}
