package cli

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monkeymonk/gdt/internal/config"
	"github.com/monkeymonk/gdt/internal/engine"
	"github.com/monkeymonk/gdt/internal/platform"
)

// TestRemove_NoArgs_NonTTY_ReturnsVersionRequiredError proves that running
// "gdt remove" with no version argument in a non-interactive process
// (isTTY() is false for the test binary's stdin) returns the "version
// required" error instead of hanging on a prompt.
func TestRemove_NoArgs_NonTTY_ReturnsVersionRequiredError(t *testing.T) {
	forceNonTTYStdin(t)
	home := t.TempDir()
	if err := os.MkdirAll(filepath.Join(home, "versions"), 0o755); err != nil {
		t.Fatal(err)
	}

	app := &App{Home: home, Platform: platform.Info{OS: "linux", Arch: "amd64"}, Config: &config.Config{}}

	cmd := newRemoveCmd(app)
	cmd.SetArgs([]string{})
	cmd.SilenceUsage = true
	cmd.SilenceErrors = true

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected an error when no version is given and no TTY is available, got nil")
	}
	if !strings.Contains(err.Error(), "version required") {
		t.Errorf("expected \"version required\" error, got: %v", err)
	}
}

// TestRemove_CurrentDefaultVersion_WarnsOnStderrAndStillRemoves proves that
// removing the version currently set as the global default prints a warning
// to stderr, and (since isTTY() is false in the test process, so the
// confirm-prompt block at remove.go's `if isTTY() { ... }` is skipped
// entirely) removal proceeds directly afterward.
func TestRemove_CurrentDefaultVersion_WarnsOnStderrAndStillRemoves(t *testing.T) {
	forceNonTTYStdin(t)
	home := t.TempDir()
	versionDir := filepath.Join(home, "versions", "4.2.1")
	if err := os.MkdirAll(versionDir, 0o755); err != nil {
		t.Fatal(err)
	}

	app := &App{
		Home:     home,
		Platform: platform.Info{OS: "linux", Arch: "amd64"},
		Config:   &config.Config{DefaultVersion: "4.2.1"},
	}

	cmd := newRemoveCmd(app)
	cmd.SetArgs([]string{"4.2.1"})
	cmd.SilenceUsage = true
	cmd.SilenceErrors = true

	origStderr := os.Stderr
	r, w, pipeErr := os.Pipe()
	if pipeErr != nil {
		t.Fatalf("os.Pipe(): %v", pipeErr)
	}
	os.Stderr = w
	t.Cleanup(func() { os.Stderr = origStderr })

	execErr := cmd.Execute()

	if err := w.Close(); err != nil {
		t.Fatalf("closing pipe writer: %v", err)
	}
	buf := make([]byte, 4096)
	n, _ := r.Read(buf)
	output := string(buf[:n])

	if execErr != nil {
		t.Fatalf("unexpected error: %v", execErr)
	}
	if !strings.Contains(output, "Warning: 4.2.1 is the current global default") {
		t.Errorf("expected a global-default warning on stderr, got: %q", output)
	}
	if !strings.Contains(output, "Version 4.2.1 removed") {
		t.Errorf("expected removal confirmation on stderr, got: %q", output)
	}
	if _, statErr := os.Stat(versionDir); !os.IsNotExist(statErr) {
		t.Errorf("expected version directory to be removed, stat err: %v", statErr)
	}
}

// TestRemove_NotInstalledVersion_PropagatesServiceError proves that when
// svc.Remove() fails (here because the requested version was never
// installed), newRemoveCmd's RunE propagates that error as-is (a
// *engine.ActionableError) rather than swallowing or wrapping it further.
//
// Note: remove.go's confirm-abort path (`if isTTY() { ok, err :=
// promptConfirm(...); ... }`) is gated entirely behind isTTY(). Forcing it
// true reliably and portably (to reach the confirm prompt at all, then
// prove the abort branch) has no test seam — promptConfirm isn't
// injectable, and this sandbox's own stdin already reporting as a real
// char-device by default is an environment artifact, not something a
// portable test should rely on (a typical headless CI runner has no
// controlling terminal at all). Not covered here — an accepted-incomplete
// gap, not a missed case.
func TestRemove_NotInstalledVersion_PropagatesServiceError(t *testing.T) {
	forceNonTTYStdin(t)
	home := t.TempDir()
	if err := os.MkdirAll(filepath.Join(home, "versions"), 0o755); err != nil {
		t.Fatal(err)
	}

	app := &App{Home: home, Platform: platform.Info{OS: "linux", Arch: "amd64"}, Config: &config.Config{}}

	cmd := newRemoveCmd(app)
	cmd.SetArgs([]string{"9.9.9"})
	cmd.SilenceUsage = true
	cmd.SilenceErrors = true

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected an error when removing a version that was never installed, got nil")
	}

	var ae *engine.ActionableError
	if !errors.As(err, &ae) {
		t.Fatalf("expected *engine.ActionableError, got %T: %v", err, err)
	}
	if !strings.Contains(err.Error(), "9.9.9 is not installed") {
		t.Errorf("expected error to mention the version is not installed, got: %v", err)
	}
	if ae.Suggestion == "" {
		t.Error("expected non-empty Suggestion on ActionableError")
	}
}
