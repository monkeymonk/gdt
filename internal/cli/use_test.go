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

func TestUse_AfterHookFailure_ReturnsError(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("skipping on windows: fake plugin binary is a #!/bin/sh script")
	}
	home := t.TempDir()
	writeFailingV2HookPlugin(t, filepath.Join(home, "plugins"), "failplugin", "after_use")

	app := &App{Home: home, Platform: platform.Info{OS: "linux", Arch: "amd64"}, Config: &config.Config{}}
	app.ConfigPath = filepath.Join(home, "config.toml")

	cmd := newUseCmd(app)
	cmd.SetArgs([]string{"4.3"})
	cmd.SilenceUsage = true
	cmd.SilenceErrors = true

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error when after_use hook fails, got nil")
	}

	var ae *engine.ActionableError
	if !errors.As(err, &ae) {
		t.Fatalf("expected *engine.ActionableError, got %T: %v", err, err)
	}
	if ae.Suggestion == "" {
		t.Error("expected non-empty Suggestion on ActionableError")
	}
	if !strings.Contains(err.Error(), "after_use") {
		t.Errorf("expected error to mention after_use hook, got: %v", err)
	}
	if !strings.Contains(err.Error(), "default version set to 4.3") {
		t.Errorf("expected error to make clear the default was set despite hook failure, got: %v", err)
	}

	// The default version must have been persisted before the hook ran,
	// proving this is a hook failure, not a use failure.
	saved, statErr := os.Stat(app.ConfigPath)
	if statErr != nil || saved.Size() == 0 {
		t.Errorf("expected config to have been saved despite hook failure: %v", statErr)
	}
	loaded, loadErr := config.Load(app.ConfigPath)
	if loadErr != nil {
		t.Fatalf("loading saved config: %v", loadErr)
	}
	if loaded.DefaultVersion != "4.3" {
		t.Errorf("expected DefaultVersion to be persisted as 4.3, got %q", loaded.DefaultVersion)
	}
}
