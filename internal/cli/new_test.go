package cli

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/monkeymonk/gdt/internal/config"
	"github.com/monkeymonk/gdt/internal/engine"
	"github.com/monkeymonk/gdt/internal/platform"
	"github.com/monkeymonk/gdt/internal/plugins"
)

// TestResolveTemplate_PluginDiscoveryFailure_ReturnsError proves that when
// plugin-template discovery fails, resolveTemplate returns the error
// instead of silently treating it as "no plugin templates available".
func TestResolveTemplate_PluginDiscoveryFailure_ReturnsError(t *testing.T) {
	pluginsDir := t.TempDir()
	// Make the plugins directory itself a file so Discover-style os.ReadDir
	// fails; pluginsDir here is the directory whose *contents* we're about
	// to seed a colliding file into, so nest one level.
	brokenPluginsDir := filepath.Join(pluginsDir, "plugins")
	if err := os.WriteFile(brokenPluginsDir, []byte("not a directory"), 0o644); err != nil {
		t.Fatalf("seeding fake plugins file: %v", err)
	}
	pluginSvc := plugins.NewService(brokenPluginsDir)

	err := resolveTemplate("mytemplate", pluginSvc, t.TempDir(), "proj", "4.3", "forward_plus", false)
	if err == nil {
		t.Fatal("expected resolveTemplate to return an error when plugin template discovery fails, got nil")
	}

	var ae *engine.ActionableError
	if !errors.As(err, &ae) {
		t.Fatalf("expected *engine.ActionableError, got %T: %v", err, err)
	}
	if ae.Suggestion == "" {
		t.Error("expected non-empty Suggestion on ActionableError")
	}
}

// TestRunNew_VersionListFailure_ReturnsError proves that when listing
// installed engine versions fails, runNew returns the error instead of
// silently proceeding with an empty (indistinguishable from "none
// installed") version list.
func TestRunNew_VersionListFailure_ReturnsError(t *testing.T) {
	home := t.TempDir()
	// Make the versions directory a regular file so svc.ListVersionStrings()
	// -> os.ReadDir(VersionsDir()) fails with a real error.
	if err := os.WriteFile(filepath.Join(home, "versions"), []byte("not a directory"), 0o644); err != nil {
		t.Fatalf("seeding fake versions file: %v", err)
	}

	app := &App{
		Home:     home,
		Config:   &config.Config{},
		Platform: platform.Info{OS: "linux", Arch: "amd64"},
	}

	// name and version both non-empty so runNew doesn't try to enter any
	// interactive huh prompt before reaching the version-list check.
	err := runNew(app, false, "proj", "", "4.3", "forward_plus", false, false, true)
	if err == nil {
		t.Fatal("expected runNew to return an error when listing installed versions fails, got nil")
	}

	var ae *engine.ActionableError
	if !errors.As(err, &ae) {
		t.Fatalf("expected *engine.ActionableError, got %T: %v", err, err)
	}
	if ae.Suggestion == "" {
		t.Error("expected non-empty Suggestion on ActionableError")
	}
}
