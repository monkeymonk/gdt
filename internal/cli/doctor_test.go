package cli

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monkeymonk/gdt/internal/config"
	"github.com/monkeymonk/gdt/internal/platform"
)

// captureStdout runs fn while redirecting os.Stdout to a pipe and returns
// everything written to it.
func captureStdout(t *testing.T, fn func()) string {
	t.Helper()

	orig := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe(): %v", err)
	}
	os.Stdout = w
	t.Cleanup(func() { os.Stdout = orig })

	fn()

	if err := w.Close(); err != nil {
		t.Fatalf("closing pipe writer: %v", err)
	}
	out, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("reading captured stdout: %v", err)
	}
	return string(out)
}

// TestRunDoctor_ListFailureSurfacesAsFailedCheck proves that when the engine
// service's List() detection call fails, runDoctor reports it as a FAILED
// check result rather than silently ignoring it (a `_` discard) or panicking.
// runDoctor itself must still return nil (doctor never aborts the whole run
// on a single detection failure).
func TestRunDoctor_ListFailureSurfacesAsFailedCheck(t *testing.T) {
	home := t.TempDir()
	// Make the versions directory a regular file instead of a directory, so
	// svc.List() -> os.ReadDir(VersionsDir()) fails with a non-NotExist
	// error (ENOTDIR), exercising the detection-failure path.
	if err := os.WriteFile(filepath.Join(home, "versions"), []byte("not a directory"), 0o644); err != nil {
		t.Fatalf("seeding fake versions file: %v", err)
	}

	app := &App{
		Home:     home,
		Config:   &config.Config{},
		Platform: platform.Info{OS: "linux", Arch: "amd64"},
	}

	var runErr error
	output := captureStdout(t, func() {
		runErr = runDoctor(app)
	})

	if runErr != nil {
		t.Fatalf("expected runDoctor to return nil (diagnostic, not fatal), got: %v", runErr)
	}
	if !strings.Contains(output, "FAIL") || !strings.Contains(output, "could not list installed engine versions") {
		t.Errorf("expected a FAILED check reporting the list failure, got output:\n%s", output)
	}
	if strings.Contains(output, "All checks passed") {
		t.Errorf("expected doctor to report issues, not a clean pass, got output:\n%s", output)
	}
}
