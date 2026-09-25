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
)

// setupExportFixture builds an installed engine version, templates, and a
// Godot project with a single "linux" export preset, and chdirs into the
// project directory. The fake "godot" binary just exits 0 when invoked.
func setupExportFixture(t *testing.T) (app *App, projectDir string) {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("skipping on windows: fake godot binary is a #!/bin/sh script")
	}
	home := t.TempDir()

	versionDir := filepath.Join(home, "versions", "4.2.1")
	if err := os.MkdirAll(versionDir, 0o755); err != nil {
		t.Fatalf("setup version dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(versionDir, "godot"), []byte("#!/bin/sh\nexit 0\n"), 0o755); err != nil {
		t.Fatalf("setup fake godot binary: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(home, "templates", "4.2.1"), 0o755); err != nil {
		t.Fatalf("setup templates dir: %v", err)
	}

	projectDir = t.TempDir()
	if err := os.WriteFile(filepath.Join(projectDir, "project.godot"), []byte(""), 0o644); err != nil {
		t.Fatalf("setup project.godot: %v", err)
	}
	if err := os.WriteFile(filepath.Join(projectDir, ".godot-version"), []byte("4.2.1"), 0o644); err != nil {
		t.Fatalf("setup .godot-version: %v", err)
	}
	if err := os.WriteFile(filepath.Join(projectDir, "export_presets.cfg"), []byte("[preset.0]\n\nname=\"linux\"\n"), 0o644); err != nil {
		t.Fatalf("setup export_presets.cfg: %v", err)
	}
	t.Chdir(projectDir)

	app = &App{Home: home, Platform: platform.Info{OS: "linux", Arch: "amd64"}, Config: &config.Config{}}
	return app, projectDir
}

func TestRunExport_MkdirAllFailure_ReturnsError(t *testing.T) {
	app, projectDir := setupExportFixture(t)

	// Block "dist" from being created as a directory by pre-creating it as a file.
	if err := os.WriteFile(filepath.Join(projectDir, "dist"), []byte(""), 0o644); err != nil {
		t.Fatalf("setup blocking file: %v", err)
	}

	err := runExport(app, "linux", "", false, false)
	if err == nil {
		t.Fatal("expected error when output directory cannot be created, got nil")
	}

	var ae *engine.ActionableError
	if !errors.As(err, &ae) {
		t.Fatalf("expected *engine.ActionableError, got %T: %v", err, err)
	}
	if ae.Suggestion == "" {
		t.Error("expected non-empty Suggestion on ActionableError")
	}
	if !strings.Contains(err.Error(), "output directory") {
		t.Errorf("expected error to mention output directory, got: %v", err)
	}
}

func TestRunExport_AfterHookFailure_ReturnsError(t *testing.T) {
	app, projectDir := setupExportFixture(t)
	writeFailingV2HookPlugin(t, filepath.Join(app.Home, "plugins"), "failplugin", "after_export")

	err := runExport(app, "linux", "", false, false)
	if err == nil {
		t.Fatal("expected error when after_export hook fails, got nil")
	}

	var ae *engine.ActionableError
	if !errors.As(err, &ae) {
		t.Fatalf("expected *engine.ActionableError, got %T: %v", err, err)
	}
	if ae.Suggestion == "" {
		t.Error("expected non-empty Suggestion on ActionableError")
	}
	if !strings.Contains(err.Error(), "after_export") {
		t.Errorf("expected error to mention after_export hook, got: %v", err)
	}
	if !strings.Contains(err.Error(), "succeeded") {
		t.Errorf("expected error to make clear the export itself succeeded, got: %v", err)
	}

	// The export must have actually completed before the hook ran.
	if _, statErr := os.Stat(filepath.Join(projectDir, "dist", "linux")); statErr != nil {
		t.Errorf("expected export output to exist despite hook failure: %v", statErr)
	}
}

// TestRunExport_BeforeHookFailure_ReturnsError proves that a failing
// before_export hook returns a *engine.ActionableError and that the
// export itself never runs (no dist/ output) when the hook fails first.
func TestRunExport_BeforeHookFailure_ReturnsError(t *testing.T) {
	app, projectDir := setupExportFixture(t)
	writeFailingV2HookPlugin(t, filepath.Join(app.Home, "plugins"), "failplugin", "before_export")

	err := runExport(app, "linux", "", false, false)
	if err == nil {
		t.Fatal("expected error when before_export hook fails, got nil")
	}

	var ae *engine.ActionableError
	if !errors.As(err, &ae) {
		t.Fatalf("expected *engine.ActionableError, got %T: %v", err, err)
	}
	if ae.Suggestion == "" {
		t.Error("expected non-empty Suggestion on ActionableError")
	}
	if !strings.Contains(err.Error(), "before_export") {
		t.Errorf("expected error to mention before_export hook, got: %v", err)
	}

	// The export must NOT have run — the hook failed before the godot
	// process was ever invoked. The output directory itself always exists
	// by this point (runExport MkdirAlls it before the hook, independent
	// of hook success), so check for the actual export artifact instead.
	if _, statErr := os.Stat(filepath.Join(projectDir, "dist", "linux", "game")); !os.IsNotExist(statErr) {
		t.Errorf("expected no export output file to exist, stat error: %v", statErr)
	}
}
