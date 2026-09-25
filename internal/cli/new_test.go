package cli

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
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

	err := resolveTemplate(resolveTemplateOptions{
		TemplateURL: "mytemplate",
		PluginSvc:   pluginSvc,
		ProjectDir:  t.TempDir(),
		Name:        "proj",
		Version:     "4.3",
		Renderer:    "forward_plus",
		CSharp:      false,
	})
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

// TestRunNew_BeforeNewHookFailure_ReturnsError proves that a failing
// before_new hook returns a *engine.ActionableError and that the project
// is never scaffolded (resolveTemplate is never reached) when the hook
// fails before creation.
func TestRunNew_BeforeNewHookFailure_ReturnsError(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("skipping on windows: fake plugin binary is a #!/bin/sh script")
	}
	home := t.TempDir()
	if err := os.MkdirAll(filepath.Join(home, "versions"), 0o755); err != nil {
		t.Fatal(err)
	}
	writeFailingV2HookPlugin(t, filepath.Join(home, "plugins"), "failplugin", "before_new")

	app := &App{
		Home:     home,
		Config:   &config.Config{},
		Platform: platform.Info{OS: "linux", Arch: "amd64"},
	}

	projectDir := t.TempDir()
	t.Chdir(projectDir)
	name := filepath.Join(projectDir, "proj")

	err := runNew(app, false, name, "", "4.3", "forward_plus", false, false, true)
	if err == nil {
		t.Fatal("expected error when before_new hook fails, got nil")
	}

	var ae *engine.ActionableError
	if !errors.As(err, &ae) {
		t.Fatalf("expected *engine.ActionableError, got %T: %v", err, err)
	}
	if ae.Suggestion == "" {
		t.Error("expected non-empty Suggestion on ActionableError")
	}
	if !strings.Contains(err.Error(), "before_new") {
		t.Errorf("expected error to mention before_new hook, got: %v", err)
	}

	// project.godot must NOT have been created — the hook failed before
	// scaffolding ran.
	if _, statErr := os.Stat(filepath.Join(name, "project.godot")); !os.IsNotExist(statErr) {
		t.Errorf("expected no project.godot to exist, stat error: %v", statErr)
	}
}

// TestRunNew_AfterNewHookFailure_ReturnsError proves that a failing
// after_new hook returns a *engine.ActionableError with phrasing that
// reflects the project WAS already created (distinct from the before_new
// case), and that the project files are on disk despite the hook failure.
func TestRunNew_AfterNewHookFailure_ReturnsError(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("skipping on windows: fake plugin binary is a #!/bin/sh script")
	}
	home := t.TempDir()
	if err := os.MkdirAll(filepath.Join(home, "versions"), 0o755); err != nil {
		t.Fatal(err)
	}
	writeFailingV2HookPlugin(t, filepath.Join(home, "plugins"), "failplugin", "after_new")

	app := &App{
		Home:     home,
		Config:   &config.Config{},
		Platform: platform.Info{OS: "linux", Arch: "amd64"},
	}

	projectDir := t.TempDir()
	t.Chdir(projectDir)
	name := filepath.Join(projectDir, "proj")

	err := runNew(app, false, name, "", "4.3", "forward_plus", false, false, true)
	if err == nil {
		t.Fatal("expected error when after_new hook fails, got nil")
	}

	var ae *engine.ActionableError
	if !errors.As(err, &ae) {
		t.Fatalf("expected *engine.ActionableError, got %T: %v", err, err)
	}
	if ae.Suggestion == "" {
		t.Error("expected non-empty Suggestion on ActionableError")
	}
	if !strings.Contains(err.Error(), "after_new") {
		t.Errorf("expected error to mention after_new hook, got: %v", err)
	}
	if !strings.Contains(err.Error(), "created successfully") {
		t.Errorf("expected error to make clear the project was created despite hook failure, got: %v", err)
	}

	// project.godot MUST exist — scaffolding succeeded before the hook ran.
	if _, statErr := os.Stat(filepath.Join(name, "project.godot")); statErr != nil {
		t.Errorf("expected project.godot to exist despite hook failure: %v", statErr)
	}
}
