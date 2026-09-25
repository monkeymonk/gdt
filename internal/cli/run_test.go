package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monkeymonk/gdt/internal/config"
	"github.com/monkeymonk/gdt/internal/platform"
)

func testApp(t *testing.T) *App {
	t.Helper()
	home := t.TempDir()
	if err := os.MkdirAll(filepath.Join(home, "versions"), 0o755); err != nil {
		t.Fatal(err)
	}
	return &App{
		Home:     home,
		Config:   &config.Config{},
		Platform: platform.Info{OS: "linux", Arch: "amd64"},
	}
}

func setupFakeInstalledVersion(t *testing.T, app *App, version string) {
	t.Helper()
	dir := filepath.Join(app.Home, "versions", version)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "godot"), []byte("bin"), 0o755); err != nil {
		t.Fatal(err)
	}
}

// TestRunGodot_ProjectRootDetectionFailure proves that when project root
// detection fails, runGodot returns a non-nil error instead of silently
// dispatching the before_run hook with a zero-value ProjectRoot.
func TestRunGodot_ProjectRootDetectionFailure(t *testing.T) {
	app := testApp(t)
	setupFakeInstalledVersion(t, app, "4.2.1")

	// Run from a directory with no project.godot in itself or any ancestor
	// up to the filesystem root, so project.DetectRoot fails.
	t.Chdir(t.TempDir())

	err := runGodot(app, []string{"4.2.1"}, false)
	if err == nil {
		t.Fatal("expected error when project root detection fails, got nil")
	}
	if !strings.Contains(err.Error(), "detecting project root") {
		t.Errorf("expected error to mention project root detection, got: %v", err)
	}
}
