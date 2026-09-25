package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monkeymonk/gdt/internal/config"
	"github.com/monkeymonk/gdt/internal/platform"
	"github.com/monkeymonk/gdt/internal/plugins"
)

// TestDispatchPlugin_FailsClosedOnProjectResolutionFailure proves that when
// the working directory does not resolve to a Godot project, dispatchPlugin
// returns the error instead of silently dispatching the plugin with a
// zero-value (empty GDT_PROJECT_ROOT) context.
func TestDispatchPlugin_FailsClosedOnProjectResolutionFailure(t *testing.T) {
	home := t.TempDir()
	// A directory with no project.godot anywhere above it in the temp tree.
	noProjectDir := t.TempDir()

	origWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("os.Getwd() before test: %v", err)
	}
	if err := os.Chdir(noProjectDir); err != nil {
		t.Fatalf("os.Chdir(%q): %v", noProjectDir, err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(origWd); err != nil {
			t.Fatalf("restoring working directory: %v", err)
		}
	})

	app := &App{
		Home:     home,
		Config:   &config.Config{},
		Platform: platform.Info{OS: "linux", Arch: "amd64"},
	}

	p := plugins.Plugin{
		Dir: t.TempDir(),
		Manifest: plugins.Manifest{
			Name: "doesnotexist",
		},
	}

	dispatchErr := dispatchPlugin(app, p, nil)
	if dispatchErr == nil {
		t.Fatal("expected dispatchPlugin to return an error when project resolution fails, got nil")
	}
	// The error must come from the resolution step (fail-closed), not from
	// attempting to exec a nonexistent plugin binary further down the path.
	if !strings.Contains(dispatchErr.Error(), "no Godot project found") {
		t.Errorf("expected error to surface the project-resolution failure, got: %v", dispatchErr)
	}
}

// TestNewRootCmd_PluginDiscoveryFailure_PrintsWarning proves that when
// plugin discovery fails, NewRootCmd prints a warning to stderr instead of
// silently registering zero plugin commands with no explanation.
func TestNewRootCmd_PluginDiscoveryFailure_PrintsWarning(t *testing.T) {
	home := t.TempDir()
	// Make the plugins directory a regular file instead of a directory, so
	// pluginSvc.Discover() -> os.ReadDir(PluginsDir()) fails.
	if err := os.WriteFile(filepath.Join(home, "plugins"), []byte("not a directory"), 0o644); err != nil {
		t.Fatalf("seeding fake plugins file: %v", err)
	}

	app := &App{
		Home:     home,
		Config:   &config.Config{},
		Platform: platform.Info{OS: "linux", Arch: "amd64"},
	}

	origStderr := os.Stderr
	r, w, pipeErr := os.Pipe()
	if pipeErr != nil {
		t.Fatalf("os.Pipe(): %v", pipeErr)
	}
	os.Stderr = w
	t.Cleanup(func() { os.Stderr = origStderr })

	NewRootCmd(app)

	if err := w.Close(); err != nil {
		t.Fatalf("closing pipe writer: %v", err)
	}
	buf := make([]byte, 4096)
	n, _ := r.Read(buf)
	output := string(buf[:n])

	if !strings.Contains(output, "warning") || !strings.Contains(output, "plugin discovery failed") {
		t.Errorf("expected a warning about plugin discovery failure on stderr, got: %q", output)
	}
}
