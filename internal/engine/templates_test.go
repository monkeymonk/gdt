package engine

import (
	"os"
	"path/filepath"
	"testing"
)

func setupFakeTemplates(t *testing.T, svc *Service, version string) {
	t.Helper()
	dir := filepath.Join(svc.templatesDir(), version)
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
	os.RemoveAll(svc.templatesDir())

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
