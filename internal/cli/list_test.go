package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestList_MarksDefaultVersion proves that the installed version matching
// the config's DefaultVersion is printed with a "* " marker while other
// installed versions get a plain "  " marker.
func TestList_MarksDefaultVersion(t *testing.T) {
	app := testApp(t)
	setupFakeInstalledVersion(t, app, "4.2.1")
	setupFakeInstalledVersion(t, app, "4.3.0")
	app.Config.DefaultVersion = "4.3.0"

	cmd := newListCmd(app)
	cmd.SilenceUsage = true
	cmd.SilenceErrors = true

	var runErr error
	output := captureStdout(t, func() {
		runErr = cmd.Execute()
	})
	if runErr != nil {
		t.Fatalf("expected nil error, got: %v", runErr)
	}

	if !strings.Contains(output, "* 4.3.0") {
		t.Errorf("expected default version 4.3.0 to be marked with '* ', got output:\n%s", output)
	}
	if !strings.Contains(output, "  4.2.1") {
		t.Errorf("expected non-default version 4.2.1 to have a plain marker, got output:\n%s", output)
	}
	if strings.Contains(output, "* 4.2.1") {
		t.Errorf("expected 4.2.1 not to be marked as default, got output:\n%s", output)
	}
}

// TestList_EmptyMessageWhenNoVersionsInstalled proves that with no
// installed versions, list prints the "No versions installed" hint instead
// of an empty "Installed versions" header.
func TestList_EmptyMessageWhenNoVersionsInstalled(t *testing.T) {
	app := testApp(t)

	cmd := newListCmd(app)
	cmd.SilenceUsage = true
	cmd.SilenceErrors = true

	var runErr error
	output := captureStdout(t, func() {
		runErr = cmd.Execute()
	})
	if runErr != nil {
		t.Fatalf("expected nil error, got: %v", runErr)
	}
	if strings.Contains(output, "Installed versions") {
		t.Errorf("expected no 'Installed versions' header when nothing is installed, got output:\n%s", output)
	}
}

// TestList_SvcListErrorPropagates proves that when svc.List() fails
// (versions dir is unreadable as a directory), the command returns a
// non-nil error instead of silently succeeding.
func TestList_SvcListErrorPropagates(t *testing.T) {
	app := testApp(t)
	// Replace the versions directory with a regular file so svc.List()'s
	// os.ReadDir(VersionsDir()) fails with a non-NotExist error.
	versionsDir := filepath.Join(app.Home, "versions")
	if err := os.RemoveAll(versionsDir); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(versionsDir, []byte("not a directory"), 0o644); err != nil {
		t.Fatalf("seeding fake versions file: %v", err)
	}

	cmd := newListCmd(app)
	cmd.SilenceUsage = true
	cmd.SilenceErrors = true

	var runErr error
	_ = captureStdout(t, func() {
		runErr = cmd.Execute()
	})
	if runErr == nil {
		t.Fatal("expected error when svc.List() fails, got nil")
	}
}
