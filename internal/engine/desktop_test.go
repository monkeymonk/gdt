package engine

import (
	"bytes"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/monkeymonk/gdt/internal/config"
	"github.com/monkeymonk/gdt/internal/platform"
)

// TestRemoveDesktop_LogsBestEffortFailuresWhenDebugEnabled proves that
// removeDesktop's best-effort file-removal failures are observable via
// slog.Warn when GDT_DEBUG=1, instead of being fully silent, while still
// remaining non-fatal (removeDesktop has no return value to fail).
//
// Desktop integration is Linux-only (see README's "Desktop Launcher
// (Linux)" section and installDesktop/removeDesktop's own GOOS guard), so
// this test only exercises anything on linux; it skips elsewhere rather
// than asserting on a no-op.
func TestRemoveDesktop_LogsBestEffortFailuresWhenDebugEnabled(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("desktop integration is linux-only")
	}

	home := t.TempDir()
	// applicationsDir()/iconShareDir() are derived from $HOME via
	// os.UserHomeDir(), so point HOME at a fresh temp dir with no
	// pre-existing .desktop/icon files — removeDesktop's os.Remove calls
	// will fail with a real (non-nil) "file does not exist" error, which
	// is exactly the best-effort failure path this test targets.
	t.Setenv("HOME", home)
	t.Setenv("XDG_DATA_HOME", "")
	t.Setenv("GDT_DEBUG", "1")

	var logBuf bytes.Buffer
	origLogger := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(&logBuf, nil)))
	t.Cleanup(func() { slog.SetDefault(origLogger) })

	svc := NewService(home, platform.Info{OS: "linux", Arch: "amd64"}, &config.Config{})
	svc.removeDesktop()

	output := logBuf.String()
	if output == "" {
		t.Fatal("expected removeDesktop to log a warning for the missing .desktop file/icon, got no log output")
	}
	if !bytes.Contains(logBuf.Bytes(), []byte("desktop:")) {
		t.Errorf("expected log output to mention the desktop best-effort failure, got: %s", output)
	}
}

// TestRemoveDesktop_NoLogWhenDebugDisabled proves the best-effort logging
// added in this round stays silent by default (GDT_DEBUG unset), matching
// removeDesktop's pre-existing "errors are silently ignored" contract for
// normal (non-debug) operation.
func TestRemoveDesktop_NoLogWhenDebugDisabled(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("desktop integration is linux-only")
	}

	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_DATA_HOME", "")
	t.Setenv("GDT_DEBUG", "")

	var logBuf bytes.Buffer
	origLogger := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(&logBuf, nil)))
	t.Cleanup(func() { slog.SetDefault(origLogger) })

	svc := NewService(home, platform.Info{OS: "linux", Arch: "amd64"}, &config.Config{})
	svc.removeDesktop()

	if logBuf.Len() != 0 {
		t.Errorf("expected no log output with GDT_DEBUG unset, got: %s", logBuf.String())
	}

	// Sanity check this test actually exercised the failure path (not a
	// vacuous pass because nothing failed): confirm no .desktop file was
	// there to remove in the first place — same setup as the DEBUG=1 test.
	if _, err := os.Stat(filepath.Join(home, ".local", "share", "applications", desktopFileName)); err == nil {
		t.Fatal("test setup invariant violated: .desktop file unexpectedly exists")
	}
}
