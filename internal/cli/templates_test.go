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

// TestTemplatesList_PrintsInstalled proves that installed template sets are
// printed under an "Installed templates" header.
func TestTemplatesList_PrintsInstalled(t *testing.T) {
	home := t.TempDir()
	if err := os.MkdirAll(filepath.Join(home, "templates", "4.2"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(home, "templates", "4.3"), 0o755); err != nil {
		t.Fatal(err)
	}

	app := &App{Home: home, Platform: platform.Info{OS: "linux", Arch: "amd64"}, Config: &config.Config{}}

	cmd := newTemplatesListCmd(app)
	cmd.SilenceUsage = true
	cmd.SilenceErrors = true

	var runErr error
	output := captureStdout(t, func() {
		runErr = cmd.Execute()
	})
	if runErr != nil {
		t.Fatalf("expected nil error, got: %v", runErr)
	}
	if !strings.Contains(output, "Installed templates") {
		t.Errorf("expected 'Installed templates' header, got output:\n%s", output)
	}
	if !strings.Contains(output, "4.2") || !strings.Contains(output, "4.3") {
		t.Errorf("expected both installed template versions listed, got output:\n%s", output)
	}
}

// TestTemplatesList_EmptyMessage proves that with no templates installed,
// list prints a "No templates installed" hint to stderr instead of an
// empty "Installed templates" header.
func TestTemplatesList_EmptyMessage(t *testing.T) {
	app := &App{Home: t.TempDir(), Platform: platform.Info{OS: "linux", Arch: "amd64"}, Config: &config.Config{}}

	cmd := newTemplatesListCmd(app)
	cmd.SilenceUsage = true
	cmd.SilenceErrors = true

	var runErr error
	output := captureStderr(t, func() {
		runErr = cmd.Execute()
	})
	if runErr != nil {
		t.Fatalf("expected nil error, got: %v", runErr)
	}
	if !strings.Contains(output, "No templates installed") {
		t.Errorf("expected 'No templates installed' hint on stderr, got:\n%s", output)
	}
}

// TestTemplatesList_SvcListErrorPropagates proves that when
// svc.ListTemplates() fails (templates dir is unreadable as a directory),
// the command returns a non-nil error instead of silently succeeding.
func TestTemplatesList_SvcListErrorPropagates(t *testing.T) {
	home := t.TempDir()
	// Seed a regular file where the templates dir needs to be a directory,
	// so svc.ListTemplates()'s os.ReadDir(TemplatesDir()) fails with a
	// non-NotExist error.
	if err := os.WriteFile(filepath.Join(home, "templates"), []byte("not a directory"), 0o644); err != nil {
		t.Fatalf("seeding fake templates file: %v", err)
	}

	app := &App{Home: home, Platform: platform.Info{OS: "linux", Arch: "amd64"}, Config: &config.Config{}}

	cmd := newTemplatesListCmd(app)
	cmd.SilenceUsage = true
	cmd.SilenceErrors = true

	var runErr error
	_ = captureStdout(t, func() {
		runErr = cmd.Execute()
	})
	if runErr == nil {
		t.Fatal("expected error when svc.ListTemplates() fails, got nil")
	}
}

// TestTemplatesRemove_NoArgs_NonTTY_VersionRequiredError proves that
// running "gdt templates remove" with no version argument in a
// non-interactive process (isTTY() forced false via forceNonTTYStdin, the
// same hermetic helper local_test.go established for this exact sandbox
// quirk) returns the "version required" error instead of hanging on a
// prompt.
func TestTemplatesRemove_NoArgs_NonTTY_VersionRequiredError(t *testing.T) {
	forceNonTTYStdin(t)
	app := &App{Home: t.TempDir(), Platform: platform.Info{OS: "linux", Arch: "amd64"}, Config: &config.Config{}}

	cmd := newTemplatesRemoveCmd(app)
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

// TestTemplatesRemove_NotInstalledVersion_PropagatesServiceError proves
// that when svc.RemoveTemplates() fails (here because the requested
// version's templates were never installed), newTemplatesRemoveCmd's RunE
// propagates that error as-is (a *engine.ActionableError) rather than
// swallowing or wrapping it further.
//
// Note: templates.go's confirm-abort path (`if isTTY() { ok, err :=
// promptConfirm(...); ... }`) is gated entirely behind isTTY(), which
// forceNonTTYStdin pins false for the duration of this test. That block is
// therefore unreachable from this test suite without a real TTY, so it is
// not covered here — an accepted-incomplete gap, matching remove_test.go's
// identical precedent for remove.go's own confirm-abort path.
func TestTemplatesRemove_NotInstalledVersion_PropagatesServiceError(t *testing.T) {
	forceNonTTYStdin(t)
	app := &App{Home: t.TempDir(), Platform: platform.Info{OS: "linux", Arch: "amd64"}, Config: &config.Config{}}

	cmd := newTemplatesRemoveCmd(app)
	cmd.SetArgs([]string{"9.9.9"})
	cmd.SilenceUsage = true
	cmd.SilenceErrors = true

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected an error when removing templates that were never installed, got nil")
	}

	var ae *engine.ActionableError
	if !errors.As(err, &ae) {
		t.Fatalf("expected *engine.ActionableError, got %T: %v", err, err)
	}
	if !strings.Contains(err.Error(), "templates for 9.9.9 are not installed") {
		t.Errorf("expected error to mention templates are not installed, got: %v", err)
	}
	if ae.Suggestion == "" {
		t.Error("expected non-empty Suggestion on ActionableError")
	}
}

// TestTemplatesInstall_AlreadyInstalled_PrintsSuccessMessageAndReturnsNil
// proves that when svc.InstallTemplates() returns engine.ErrAlreadyInstalled
// (the destination templates dir for the resolved version already exists),
// the install command treats this as a non-error success outcome: RunE
// returns nil and a distinct "already installed" message is printed,
// instead of surfacing the sentinel as a command failure.
func TestTemplatesInstall_AlreadyInstalled_PrintsSuccessMessageAndReturnsNil(t *testing.T) {
	home := t.TempDir()
	apiURL := testInstallServer(t, "9.9.9")

	// Pre-create the destination templates dir so downloadAndInstall's
	// already-installed check (which runs before artifact resolution)
	// short-circuits with engine.ErrAlreadyInstalled.
	if err := os.MkdirAll(filepath.Join(home, "templates", "9.9.9"), 0o755); err != nil {
		t.Fatal(err)
	}

	app := &App{
		Home:     home,
		Platform: platform.Info{OS: "linux", Arch: "amd64"},
		Config:   &config.Config{GodotAPI: apiURL},
	}

	cmd := newTemplatesInstallCmd(app)
	cmd.SetArgs([]string{"9.9.9"})
	cmd.SilenceUsage = true
	cmd.SilenceErrors = true

	var runErr error
	output := captureStderr(t, func() {
		runErr = cmd.Execute()
	})
	if runErr != nil {
		t.Fatalf("expected nil error for the already-installed branch, got: %v", runErr)
	}
	if !strings.Contains(output, "Templates for 9.9.9 already installed") {
		t.Errorf("expected already-installed message on stderr, got:\n%s", output)
	}
}

// TestTemplatesInstall_GenericError_ReturnsNonNilError proves that a
// genuine install error is distinct from the ErrAlreadyInstalled branch:
// RunE returns a non-nil error and the "already installed" message is
// never printed. testInstallServer serves an engine-artifact release (no
// export-templates asset), and the destination templates dir does not
// pre-exist, so this hits artifact resolution's genuine "templates not
// found" error rather than the already-installed short-circuit.
func TestTemplatesInstall_GenericError_ReturnsNonNilError(t *testing.T) {
	home := t.TempDir()
	apiURL := testInstallServer(t, "8.8.8")

	app := &App{
		Home:     home,
		Platform: platform.Info{OS: "linux", Arch: "amd64"},
		Config:   &config.Config{GodotAPI: apiURL},
	}

	cmd := newTemplatesInstallCmd(app)
	cmd.SetArgs([]string{"8.8.8"})
	cmd.SilenceUsage = true
	cmd.SilenceErrors = true

	var runErr error
	output := captureStderr(t, func() {
		runErr = cmd.Execute()
	})
	if runErr == nil {
		t.Fatal("expected a non-nil error for a genuine install failure, got nil")
	}
	if !strings.Contains(runErr.Error(), "templates not found for version 8.8.8") {
		t.Errorf("expected 'templates not found' error, got: %v", runErr)
	}
	if strings.Contains(output, "already installed") {
		t.Errorf("expected no 'already installed' message on a genuine error, got:\n%s", output)
	}
}
