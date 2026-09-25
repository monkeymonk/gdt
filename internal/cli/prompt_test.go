package cli

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/monkeymonk/gdt/internal/config"
	"github.com/monkeymonk/gdt/internal/engine"
)

// TestPromptVersion_ServiceError proves that a failure listing installed
// versions is surfaced to the caller instead of being treated as "no
// versions installed" (empty slice with a swallowed error).
func TestPromptVersion_ServiceError(t *testing.T) {
	home := t.TempDir()
	// Put a regular file where the versions directory is expected, so
	// os.ReadDir fails with something other than "not exist".
	if err := os.WriteFile(filepath.Join(home, "versions"), []byte("x"), 0644); err != nil {
		t.Fatalf("setup: %v", err)
	}
	app := &App{Home: home, Config: &config.Config{}}

	version, err := promptVersion(app, "Pick a version")
	if err == nil {
		t.Fatal("expected an error, got nil")
	}
	if version != "" {
		t.Errorf("expected empty version on error, got %q", version)
	}
	var ae *engine.ActionableError
	if !errors.As(err, &ae) {
		t.Fatalf("expected *engine.ActionableError, got %T: %v", err, err)
	}
	if ae.Suggestion == "" {
		t.Error("expected a non-empty suggestion")
	}
}

// TestPromptInstalledPlugin_ServiceError proves that a failure discovering
// installed plugins is surfaced to the caller instead of being treated as
// "no plugins installed".
func TestPromptInstalledPlugin_ServiceError(t *testing.T) {
	home := t.TempDir()
	// Put a regular file where the plugins directory is expected, so
	// os.ReadDir fails with something other than "not exist".
	if err := os.WriteFile(filepath.Join(home, "plugins"), []byte("x"), 0644); err != nil {
		t.Fatalf("setup: %v", err)
	}
	app := &App{Home: home, Config: &config.Config{}}

	name, err := promptInstalledPlugin(app, "Pick a plugin")
	if err == nil {
		t.Fatal("expected an error, got nil")
	}
	if name != "" {
		t.Errorf("expected empty name on error, got %q", name)
	}
	var ae *engine.ActionableError
	if !errors.As(err, &ae) {
		t.Fatalf("expected *engine.ActionableError, got %T: %v", err, err)
	}
	if ae.Suggestion == "" {
		t.Error("expected a non-empty suggestion")
	}
}

// TestPromptInstalledTemplate_ServiceError proves that a failure listing
// installed templates is surfaced to the caller instead of being treated
// as "no templates installed".
func TestPromptInstalledTemplate_ServiceError(t *testing.T) {
	home := t.TempDir()
	// Put a regular file where the templates directory is expected, so
	// os.ReadDir fails with something other than "not exist".
	if err := os.WriteFile(filepath.Join(home, "templates"), []byte("x"), 0644); err != nil {
		t.Fatalf("setup: %v", err)
	}
	app := &App{Home: home, Config: &config.Config{}}

	name, err := promptInstalledTemplate(app, "Pick a template")
	if err == nil {
		t.Fatal("expected an error, got nil")
	}
	if name != "" {
		t.Errorf("expected empty name on error, got %q", name)
	}
	var ae *engine.ActionableError
	if !errors.As(err, &ae) {
		t.Fatalf("expected *engine.ActionableError, got %T: %v", err, err)
	}
	if ae.Suggestion == "" {
		t.Error("expected a non-empty suggestion")
	}
}
