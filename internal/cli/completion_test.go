package cli

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/monkeymonk/gdt/internal/config"
	"github.com/monkeymonk/gdt/internal/platform"
)

// writeFailingCompletionPlugin installs a plugin under pluginsDir that
// declares `completions = true` and exits non-zero when invoked with
// "completions <shell>" args, exercising the completion command's
// per-plugin failure path (see plugins.RunPluginSubcommand call site in
// completion.go).
func writeFailingCompletionPlugin(t *testing.T, pluginsDir, name string) {
	t.Helper()
	pluginDir := filepath.Join(pluginsDir, name)
	if err := os.MkdirAll(pluginDir, 0o755); err != nil {
		t.Fatalf("setup plugin dir: %v", err)
	}
	manifest := `name = "` + name + `"
version = "1.0.0"
protocol = 2

[contributions]
completions = true
`
	if err := os.WriteFile(filepath.Join(pluginDir, "plugin.toml"), []byte(manifest), 0o644); err != nil {
		t.Fatalf("setup plugin manifest: %v", err)
	}
	binPath := filepath.Join(pluginDir, name)
	if err := os.WriteFile(binPath, []byte("#!/bin/sh\nexit 1\n"), 0o755); err != nil {
		t.Fatalf("setup plugin binary: %v", err)
	}
}

// TestCompletion_PluginFailure_WarnsAndContinues proves that when a plugin's
// `completions <shell>` subcommand fails, "gdt completion <shell>" prints a
// warning to stderr and continues rather than aborting the whole command:
// the built-in shell completion script must still reach stdout intact.
func TestCompletion_PluginFailure_WarnsAndContinues(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("skipping on windows: fake plugin binary is a #!/bin/sh script")
	}

	home := t.TempDir()
	writeFailingCompletionPlugin(t, filepath.Join(home, "plugins"), "failcompletion")

	app := &App{Home: home, Platform: platform.Info{OS: "linux", Arch: "amd64"}, Config: &config.Config{}}

	cmd := newCompletionCmd(app)
	cmd.SetArgs([]string{"bash"})
	cmd.SilenceUsage = true
	cmd.SilenceErrors = true

	origStderr := os.Stderr
	r, w, pipeErr := os.Pipe()
	if pipeErr != nil {
		t.Fatalf("os.Pipe(): %v", pipeErr)
	}
	os.Stderr = w
	t.Cleanup(func() { os.Stderr = origStderr })

	var runErr error
	stdout := captureStdout(t, func() {
		runErr = cmd.Execute()
	})

	if err := w.Close(); err != nil {
		t.Fatalf("closing pipe writer: %v", err)
	}
	buf := make([]byte, 4096)
	n, _ := r.Read(buf)
	stderr := string(buf[:n])

	if runErr != nil {
		t.Fatalf("expected completion command to return nil despite plugin failure, got: %v", runErr)
	}
	if !strings.Contains(stderr, "warning") || !strings.Contains(stderr, "failcompletion") {
		t.Errorf("expected a warning about the failed plugin completion on stderr, got: %q", stderr)
	}
	if !strings.Contains(stdout, "-*- shell-script -*-") || len(stdout) < 1000 {
		t.Errorf("expected the built-in bash completion script on stdout to be intact, got %d bytes: %q", len(stdout), stdout)
	}
}
