package plugins

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig_ExistingFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "plugin.toml")
	if err := os.WriteFile(path, []byte("[settings]\nkey = \"value\"\n"), 0644); err != nil {
		t.Fatal(err)
	}

	svc := NewService(dir)
	cfg, err := svc.LoadConfig(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	settings, ok := cfg["settings"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected settings map, got %T", cfg["settings"])
	}
	if settings["key"] != "value" {
		t.Errorf("expected key=value, got %v", settings["key"])
	}
}

func TestLoadConfig_MissingFile(t *testing.T) {
	svc := NewService(t.TempDir())
	cfg, err := svc.LoadConfig("/nonexistent/path/plugin.toml")
	if err != nil {
		t.Fatalf("expected nil error for missing file, got: %v", err)
	}
	if len(cfg) != 0 {
		t.Errorf("expected empty config, got %v", cfg)
	}
}

func TestLoadConfig_InvalidTOML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.toml")
	if err := os.WriteFile(path, []byte("not = valid = toml\n"), 0644); err != nil {
		t.Fatal(err)
	}

	svc := NewService(dir)
	_, err := svc.LoadConfig(path)
	if err == nil {
		t.Fatal("expected error for invalid TOML, got nil")
	}
}
