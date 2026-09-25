package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/monkeymonk/gdt/internal/config"
	"github.com/monkeymonk/gdt/internal/platform"
)

// TestNewApp_TrimsVersionVPrefix verifies that NewApp trims a "v" prefix
// from the version string (e.g., "v1.2.3" becomes "1.2.3").
func TestNewApp_TrimsVersionVPrefix(t *testing.T) {
	home := t.TempDir()
	t.Setenv("GDT_HOME", home)

	app, err := NewApp("v1.2.3")
	if err != nil {
		t.Fatalf("NewApp: %v", err)
	}
	if app.Version != "1.2.3" {
		t.Errorf("expected version 1.2.3, got %q", app.Version)
	}
}

// TestNewApp_ConfigLoadErrorPropagates verifies that when config.Load fails
// (e.g., because the config path is a directory instead of a file),
// NewApp returns a non-nil error.
func TestNewApp_ConfigLoadErrorPropagates(t *testing.T) {
	home := t.TempDir()
	t.Setenv("GDT_HOME", home)

	// Create config.toml as a directory, so config.Load will fail
	// when trying to os.ReadFile it.
	configDir := filepath.Join(home, "config.toml")
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatalf("setup: creating config dir: %v", err)
	}

	_, err := NewApp("v1.0.0")
	if err == nil {
		t.Errorf("expected NewApp to return an error when config.toml is a directory, got nil")
	}
}

// TestApp_PluginsDir verifies that PluginsDir returns the correct path
// by joining the home directory with "plugins".
func TestApp_PluginsDir(t *testing.T) {
	home := t.TempDir()
	app := &App{Home: home}

	pluginsDir := app.PluginsDir()
	expected := filepath.Join(home, "plugins")
	if pluginsDir != expected {
		t.Errorf("expected %q, got %q", expected, pluginsDir)
	}
}

// TestApp_EngineSvc verifies that EngineSvc returns a non-nil *engine.Service.
func TestApp_EngineSvc(t *testing.T) {
	app := &App{
		Home:     t.TempDir(),
		Platform: platform.Info{OS: "linux", Arch: "amd64"},
		Config:   &config.Config{},
	}

	svc := app.EngineSvc()
	if svc == nil {
		t.Errorf("expected EngineSvc to return a non-nil *engine.Service")
	}
}

// TestApp_PluginSvc verifies that PluginSvc lazily constructs the plugin
// service once and returns the exact same instance on every subsequent
// call, so plugin discovery caching is shared across all call sites
// within one App/process invocation.
func TestApp_PluginSvc(t *testing.T) {
	app := &App{Home: t.TempDir()}

	first := app.PluginSvc()
	if first == nil {
		t.Fatalf("expected PluginSvc to return a non-nil *plugins.Service")
	}

	second := app.PluginSvc()
	if first != second {
		t.Errorf("expected PluginSvc to return the same instance on repeated calls, got %p and %p", first, second)
	}
}
