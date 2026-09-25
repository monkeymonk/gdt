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

// writeFailingV2HookPlugin installs a V2-protocol plugin under pluginsDir that
// declares the given hook event and exits non-zero when invoked, so
// RunHooks(event, ...) returns an error.
func writeFailingV2HookPlugin(t *testing.T, pluginsDir, name, event string) {
	t.Helper()
	pluginDir := filepath.Join(pluginsDir, name)
	if err := os.MkdirAll(pluginDir, 0o755); err != nil {
		t.Fatalf("setup plugin dir: %v", err)
	}
	manifest := `name = "` + name + `"
version = "1.0.0"
protocol = 2

[contributions]
hooks = ["` + event + `"]
`
	if err := os.WriteFile(filepath.Join(pluginDir, "plugin.toml"), []byte(manifest), 0o644); err != nil {
		t.Fatalf("setup plugin manifest: %v", err)
	}
	binPath := filepath.Join(pluginDir, name)
	if err := os.WriteFile(binPath, []byte("#!/bin/sh\nexit 1\n"), 0o755); err != nil {
		t.Fatalf("setup plugin binary: %v", err)
	}
}

func TestRunCiSetup_AfterHookFailure_ReturnsError(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("skipping on windows: fake plugin binary is a #!/bin/sh script")
	}
	home := t.TempDir()
	writeFailingV2HookPlugin(t, filepath.Join(home, "plugins"), "failplugin", "after_ci_setup")

	projectDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(projectDir, "project.godot"), []byte(""), 0o644); err != nil {
		t.Fatalf("setup project.godot: %v", err)
	}
	t.Chdir(projectDir)

	app := &App{Home: home, Platform: platform.Info{OS: "linux", Arch: "amd64"}, Config: &config.Config{}}

	err := runCiSetup(app, "generic")
	if err == nil {
		t.Fatal("expected error when after_ci_setup hook fails, got nil")
	}

	var ae *engine.ActionableError
	if !errors.As(err, &ae) {
		t.Fatalf("expected *engine.ActionableError, got %T: %v", err, err)
	}
	if ae.Suggestion == "" {
		t.Error("expected non-empty Suggestion on ActionableError")
	}
	if !strings.Contains(err.Error(), "after_ci_setup") {
		t.Errorf("expected error to mention after_ci_setup hook, got: %v", err)
	}

	// The CI file itself must have been written before the hook ran, proving
	// this is a hook-failure, not a CI-setup failure.
	if _, statErr := os.Stat(filepath.Join(projectDir, "ci", "export.sh")); statErr != nil {
		t.Errorf("expected CI configuration to have been written despite hook failure: %v", statErr)
	}
}

// writeV2HookAndCIProviderPlugin installs a V2-protocol plugin that
// contributes BOTH a CI provider file and a hook that fails, so
// runCiSetup's "plugin:"-prefixed branch (ci.go:69-114) can be exercised
// end to end, including its own cwd/DetectRoot/RunHooks error handling —
// which is a separate code path from the built-in-provider branch that
// TestRunCiSetup_AfterHookFailure_ReturnsError above already covers.
func writeV2HookAndCIProviderPlugin(t *testing.T, pluginsDir, name, providerName, event string) {
	t.Helper()
	pluginDir := filepath.Join(pluginsDir, name)
	if err := os.MkdirAll(filepath.Join(pluginDir, "ci"), 0o755); err != nil {
		t.Fatalf("setup plugin ci dir: %v", err)
	}
	manifest := `name = "` + name + `"
version = "1.0.0"
protocol = 2

[contributions]
hooks = ["` + event + `"]
ci_providers = ["` + providerName + `"]
`
	if err := os.WriteFile(filepath.Join(pluginDir, "plugin.toml"), []byte(manifest), 0o644); err != nil {
		t.Fatalf("setup plugin manifest: %v", err)
	}
	if err := os.WriteFile(filepath.Join(pluginDir, "ci", providerName+".yml"), []byte("name: fake-ci\n"), 0o644); err != nil {
		t.Fatalf("setup plugin ci file: %v", err)
	}
	binPath := filepath.Join(pluginDir, name)
	if err := os.WriteFile(binPath, []byte("#!/bin/sh\nexit 1\n"), 0o755); err != nil {
		t.Fatalf("setup plugin binary: %v", err)
	}
}

func TestRunCiSetup_PluginProvider_AfterHookFailure_ReturnsError(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("skipping on windows: fake plugin binary is a #!/bin/sh script")
	}
	home := t.TempDir()
	writeV2HookAndCIProviderPlugin(t, filepath.Join(home, "plugins"), "ciplugin", "myprovider", "after_ci_setup")

	projectDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(projectDir, "project.godot"), []byte(""), 0o644); err != nil {
		t.Fatalf("setup project.godot: %v", err)
	}
	t.Chdir(projectDir)

	app := &App{Home: home, Platform: platform.Info{OS: "linux", Arch: "amd64"}, Config: &config.Config{}}

	err := runCiSetup(app, "plugin:ciplugin:myprovider")
	if err == nil {
		t.Fatal("expected error when after_ci_setup hook fails on the plugin-provider branch, got nil")
	}

	var ae *engine.ActionableError
	if !errors.As(err, &ae) {
		t.Fatalf("expected *engine.ActionableError, got %T: %v", err, err)
	}
	if ae.Suggestion == "" {
		t.Error("expected non-empty Suggestion on ActionableError")
	}
	if !strings.Contains(err.Error(), "after_ci_setup") {
		t.Errorf("expected error to mention after_ci_setup hook, got: %v", err)
	}

	// The plugin-provided CI file must have been written before the hook
	// ran, proving this is a hook failure, not a CI-setup failure — and
	// proving this test actually exercised the plugin: branch's own write
	// path (.ci/, distinct from the built-in branch's ci/ directory).
	if _, statErr := os.Stat(filepath.Join(projectDir, ".ci", "myprovider.yml")); statErr != nil {
		t.Errorf("expected plugin CI configuration to have been written despite hook failure: %v", statErr)
	}
}
