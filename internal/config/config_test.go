package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadMissingFile(t *testing.T) {
	cfg, err := Load("/nonexistent/path/config.toml")
	if err != nil {
		t.Fatal("missing config should not error")
	}
	if cfg.DefaultVersion != "" {
		t.Error("default version should be empty")
	}
}

func TestLoadValidFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")
	os.WriteFile(path, []byte(`default_version = "4.3"`), 0644)

	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.DefaultVersion != "4.3" {
		t.Errorf("DefaultVersion = %q, want %q", cfg.DefaultVersion, "4.3")
	}
}

func TestLoadInvalidToml(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")
	os.WriteFile(path, []byte(`[invalid`), 0644)

	_, err := Load(path)
	if err == nil {
		t.Error("should error on invalid TOML")
	}
}

func TestSave(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")

	cfg := &Config{DefaultVersion: "4.2"}
	err := Save(path, cfg)
	if err != nil {
		t.Fatal(err)
	}

	loaded, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.DefaultVersion != "4.2" {
		t.Errorf("DefaultVersion = %q, want %q", loaded.DefaultVersion, "4.2")
	}
}

func TestSaveCreatesParentDirs(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sub", "dir", "config.toml")

	cfg := &Config{DefaultVersion: "4.1"}
	err := Save(path, cfg)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Error("file should have been created")
	}
}

func TestConfig_DefaultGodotAPIURL(t *testing.T) {
	cfg := &Config{}
	if cfg.GodotAPIURL() != "https://api.github.com/repos/godotengine/godot/releases" {
		t.Errorf("unexpected default: %s", cfg.GodotAPIURL())
	}
}

func TestConfig_DefaultSelfUpdateURL(t *testing.T) {
	cfg := &Config{}
	if cfg.SelfUpdateAPIURL() != "https://api.github.com/repos/monkeymonk/gdt/releases/latest" {
		t.Errorf("unexpected default: %s", cfg.SelfUpdateAPIURL())
	}
}

func TestConfig_CustomGodotAPIURL(t *testing.T) {
	cfg := &Config{GodotAPI: "https://custom.api/releases"}
	if cfg.GodotAPIURL() != "https://custom.api/releases" {
		t.Errorf("unexpected: %s", cfg.GodotAPIURL())
	}
}

