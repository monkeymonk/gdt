package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monkeymonk/gdt/internal/config"
	"github.com/monkeymonk/gdt/internal/platform"
)

// forceNonTTYStdin redirects os.Stdin to a pipe's read end for the duration
// of the test, so isTTY() reliably reports false regardless of whether the
// test binary's own stdin happens to be attached to a real terminal (as it
// can be under some interactive sandboxes). The original os.Stdin is
// restored via t.Cleanup.
func forceNonTTYStdin(t *testing.T) {
	t.Helper()
	orig := os.Stdin
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe(): %v", err)
	}
	if err := w.Close(); err != nil {
		t.Fatalf("closing pipe writer: %v", err)
	}
	os.Stdin = r
	t.Cleanup(func() {
		os.Stdin = orig
		r.Close()
	})
}

// TestLocal_NoArgs_NonTTY_ReturnsVersionRequiredError proves that running
// "gdt local" with no version argument in a non-interactive process (isTTY()
// is false for the test binary's stdin) returns the "version required"
// error instead of hanging on a prompt.
func TestLocal_NoArgs_NonTTY_ReturnsVersionRequiredError(t *testing.T) {
	forceNonTTYStdin(t)
	home := t.TempDir()
	if err := os.MkdirAll(filepath.Join(home, "versions"), 0o755); err != nil {
		t.Fatal(err)
	}

	app := &App{Home: home, Platform: platform.Info{OS: "linux", Arch: "amd64"}, Config: &config.Config{}}

	cmd := newLocalCmd(app)
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

// TestLocal_PinsInstalledVersion_WritesResolvedVersionToFile proves that a
// successful "gdt local <query>" resolves the query against installed
// versions (here a prefix match) and writes the *resolved* version, not the
// raw query, into a .godot-version file in the current directory.
func TestLocal_PinsInstalledVersion_WritesResolvedVersionToFile(t *testing.T) {
	home := t.TempDir()
	if err := os.MkdirAll(filepath.Join(home, "versions", "4.2.1"), 0o755); err != nil {
		t.Fatal(err)
	}

	projectDir := t.TempDir()
	t.Chdir(projectDir)

	app := &App{Home: home, Platform: platform.Info{OS: "linux", Arch: "amd64"}, Config: &config.Config{}}

	cmd := newLocalCmd(app)
	cmd.SetArgs([]string{"4.2"})
	cmd.SilenceUsage = true
	cmd.SilenceErrors = true

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(projectDir, ".godot-version"))
	if err != nil {
		t.Fatalf("reading .godot-version: %v", err)
	}
	if got := string(data); got != "4.2.1\n" {
		t.Errorf("expected .godot-version to contain resolved version %q, got %q", "4.2.1\n", got)
	}
}

// TestLocal_NotInstalledVersion_WarnsAndPinsRawQuery proves that pinning a
// version that is not installed prints a "not installed" warning to stderr
// (mirroring use.go's soft-fallback) while still writing the raw, unresolved
// query into .godot-version.
func TestLocal_NotInstalledVersion_WarnsAndPinsRawQuery(t *testing.T) {
	home := t.TempDir()
	if err := os.MkdirAll(filepath.Join(home, "versions"), 0o755); err != nil {
		t.Fatal(err)
	}

	projectDir := t.TempDir()
	t.Chdir(projectDir)

	app := &App{Home: home, Platform: platform.Info{OS: "linux", Arch: "amd64"}, Config: &config.Config{}}

	cmd := newLocalCmd(app)
	cmd.SetArgs([]string{"9.9.9"})
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
	if !strings.Contains(output, "Warning: version 9.9.9 is not installed") {
		t.Errorf("expected a not-installed warning on stderr, got: %q", output)
	}

	data, err := os.ReadFile(filepath.Join(projectDir, ".godot-version"))
	if err != nil {
		t.Fatalf("reading .godot-version: %v", err)
	}
	if got := string(data); got != "9.9.9\n" {
		t.Errorf("expected .godot-version to contain the raw query %q, got %q", "9.9.9\n", got)
	}
}
